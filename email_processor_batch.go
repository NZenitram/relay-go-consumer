package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"relay-go-consumer/database"
	"time"

	"github.com/IBM/sarama"
)

func ProcessBatchEmails(msg *sarama.ConsumerMessage) {
	var batchEmail struct {
		BatchID         int
		Personalization Personalization
		From            EmailAddress
		Content         []Content
		Attachments     []Attachment
		Headers         map[string]string
		Sections        map[string]string
		Categories      []string
	}

	err := json.Unmarshal(msg.Value, &batchEmail)
	if err != nil {
		log.Printf("Failed to parse batch email JSON: %v", err)
		return
	}

	database.InitDB()
	db := database.GetDB()

	// Fetch batch info including created_at and initial_weights
	var batchInfo struct {
		CreatedAt      time.Time
		InitialWeights map[string]int
	}
	err = db.QueryRow("SELECT created_at, initial_weights FROM email_batches WHERE id = $1", batchEmail.BatchID).Scan(
		&batchInfo.CreatedAt, &batchInfo.InitialWeights)
	if err != nil {
		log.Printf("Failed to fetch batch info: %v", err)
		return
	}

	// Fetch ESP credentials
	credentials, err := fetchESPCredentials(batchEmail.BatchID)
	if err != nil {
		log.Printf("Failed to fetch ESP credentials: %v", err)
		return
	}

	// Calculate new weights based on recent events
	newWeights, err := calculateRecentWeights(db, batchEmail.BatchID, batchInfo.CreatedAt)
	if err != nil {
		log.Printf("Failed to calculate recent weights: %v", err)
		// Fall back to initial weights if there's an error
		newWeights = batchInfo.InitialWeights
	}

	// Compare new weights with initial weights and adjust
	adjustedWeights := adjustWeights(batchInfo.InitialWeights, newWeights)

	// Construct email message
	emailMessage := EmailMessage{
		From:             batchEmail.From,
		To:               []EmailAddress{batchEmail.Personalization.To},
		Subject:          batchEmail.Personalization.Subject,
		Content:          batchEmail.Content,
		Attachments:      batchEmail.Attachments,
		Headers:          batchEmail.Headers,
		Sections:         batchEmail.Sections,
		Categories:       batchEmail.Categories,
		Personalizations: []Personalization{batchEmail.Personalization},
		Credentials:      credentials,
	}

	// Select sender based on adjusted weights
	sender := SelectSender(adjustedWeights)

	// Send email
	switch sender {
	case "sendgrid":
		SendEmailWithSendGrid(emailMessage)
	case "socketlabs":
		SendEmailWithSocketLabs(emailMessage)
	case "postmark":
		SendEmailWithPostmark(emailMessage)
	case "sparkpost":
		SendEmailWithSparkPost(emailMessage)
	default:
		log.Printf("No valid credentials found for sender: %s", sender)
	}

	// Update batch status
	err = updateBatchStatus(db, batchEmail.BatchID)
	if err != nil {
		log.Printf("Failed to update batch status: %v", err)
	}

	// Optionally, you could update the weights in the database here if you want to persist the adjusted weights
	// updateBatchWeights(db, batchEmail.BatchID, adjustedWeights)
}

func handleBatchSend(db *sql.DB, userID int, emailMessage EmailMessage) error {
	// Calculate initial weights
	credentials, err := fetchESPCredentials(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch ESP credentials: %v", err)
	}

	initialWeights, err := calculateWeights(db, userID, credentials)
	if err != nil {
		return fmt.Errorf("failed to calculate initial weights: %v", err)
	}

	// Create a batch record in the database
	batchInfo, err := createBatchRecord(db, userID, len(emailMessage.Personalizations), initialWeights)
	if err != nil {
		return fmt.Errorf("failed to create batch record: %v", err)
	}

	// Queue emails for batch processing
	err = queueEmailsForBatch(batchInfo.ID, emailMessage)
	if err != nil {
		return fmt.Errorf("failed to queue emails for batch: %v", err)
	}

	return nil
}

func createBatchRecord(db *sql.DB, userID int, totalEmails int, initialWeights map[string]int) (*BatchInfo, error) {
	batchInfo := &BatchInfo{
		UserID:         userID,
		TotalEmails:    totalEmails,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Status:         "pending",
		InitialWeights: initialWeights,
	}

	weightsJSON, err := json.Marshal(initialWeights)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal initial weights: %v", err)
	}

	err = db.QueryRow(`
        INSERT INTO email_batches (user_id, total_messages, created_at, updated_at, status, initial_weights)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `, batchInfo.UserID, batchInfo.TotalEmails, batchInfo.CreatedAt, batchInfo.UpdatedAt, batchInfo.Status, weightsJSON).Scan(&batchInfo.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to insert batch record: %v", err)
	}

	return batchInfo, nil
}

func queueEmailsForBatch(batchID int, emailMessage EmailMessage) error {
	kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}
	if len(kafkaBrokers) == 0 {
		log.Fatal("KAFKA_BROKERS environment variable is not set")
	}
	// Set up the Kafka producer
	producer, err := sarama.NewSyncProducer(kafkaBrokers, nil)
	if err != nil {
		log.Fatalf("Failed to start Kafka producer: %v", err)
	}
	defer producer.Close()

	for _, p := range emailMessage.Personalizations {
		batchEmail := struct {
			BatchID         int
			Personalization Personalization
			From            EmailAddress
			Content         []Content
			Attachments     []Attachment
			Headers         map[string]string
			Sections        map[string]string
			Categories      []string
		}{
			BatchID:         batchID,
			Personalization: p,
			From:            emailMessage.From,
			Content:         emailMessage.Content,
			Attachments:     emailMessage.Attachments,
			Headers:         emailMessage.Headers,
			Sections:        emailMessage.Sections,
			Categories:      emailMessage.Categories,
		}

		batchEmailJSON, err := json.Marshal(batchEmail)
		if err != nil {
			return err
		}

		_, _, err = producer.SendMessage(&sarama.ProducerMessage{
			Topic: "batch_emails",
			Key:   sarama.StringEncoder(fmt.Sprintf("%d", batchID)),
			Value: sarama.StringEncoder(batchEmailJSON),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func updateBatchStatus(db *sql.DB, batchID int) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Get the current batch information
	var totalEmails, processedEmails int
	var currentStatus string
	err = tx.QueryRow(`
        SELECT total_emails, processed_emails, status 
        FROM email_batches 
        WHERE id = $1 
        FOR UPDATE
    `, batchID).Scan(&totalEmails, &processedEmails, &currentStatus)
	if err != nil {
		return fmt.Errorf("failed to get batch info: %v", err)
	}

	// Increment the processed emails count
	processedEmails++

	// Determine the new status
	var newStatus string
	if processedEmails >= totalEmails {
		newStatus = "completed"
	} else if currentStatus == "pending" {
		newStatus = "processing"
	} else {
		newStatus = currentStatus
	}

	// Update the batch status
	_, err = tx.Exec(`
        UPDATE email_batches 
        SET processed_emails = $1, status = $2 
        WHERE id = $3
    `, processedEmails, newStatus, batchID)
	if err != nil {
		return fmt.Errorf("failed to update batch status: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
