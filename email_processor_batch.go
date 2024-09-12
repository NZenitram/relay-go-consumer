package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"relay-go-consumer/database"
	"strconv"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// func ProcessBatchEmails(msg *sarama.ConsumerMessage) {
// 	var batchEmail struct {
// 		BatchID         int
// 		Personalization Personalization
// 		From            EmailAddress
// 		Content         []Content
// 		Attachments     []Attachment
// 		Headers         map[string]string
// 		Sections        map[string]string
// 		Categories      []string
// 	}

// 	err := json.Unmarshal(msg.Value, &batchEmail)
// 	if err != nil {
// 		log.Printf("Failed to parse batch email JSON: %v", err)
// 		return
// 	}

// 	database.InitDB()
// 	db := database.GetDB()

// 	// Fetch batch info including created_at and initial_weights
// 	var batchInfo struct {
// 		UserID          int
// 		CreatedAt       time.Time
// 		IntervalSeconds int
// 		InitialWeights  string
// 	}
// 	err = db.QueryRow("SELECT user_id, created_at, initial_weights FROM email_batches WHERE id = $1", batchEmail.BatchID).Scan(
// 		&batchInfo.UserID, &batchInfo.CreatedAt, &batchInfo.InitialWeights)
// 	if err != nil {
// 		log.Printf("Failed to fetch batch info: %v", err)
// 		return
// 	}

// 	var initialWeights map[string]int
// 	err = json.Unmarshal([]byte(batchInfo.InitialWeights), &initialWeights)
// 	if err != nil {
// 		log.Printf("Failed to unmarshal initial weights: %v", err)
// 		return
// 	}

// 	// Fetch ESP credentials
// 	credentials, err := fetchESPCredentials(batchInfo.UserID)
// 	if err != nil {
// 		log.Printf("Failed to fetch ESP credentials: %v", err)
// 		return
// 	}

// 	// Calculate new weights based on recent events
// 	newWeights, err := calculateRecentWeights(db, batchEmail.BatchID, batchInfo.CreatedAt)
// 	if err != nil {
// 		log.Printf("Failed to calculate recent weights for BatchID: %v - CreatedAt: %v, With Error: %v", batchEmail.BatchID, batchInfo.CreatedAt, err)
// 		// Fall back to initial weights if there's an error
// 		newWeights = initialWeights
// 	}

// 	// Compare new weights with initial weights and adjust
// 	adjustedWeights := adjustWeights(initialWeights, newWeights)

// 	// Construct email message
// 	emailMessage := EmailMessage{
// 		From:             batchEmail.From,
// 		To:               []EmailAddress{batchEmail.Personalization.To},
// 		Subject:          batchEmail.Personalization.Subject,
// 		Content:          batchEmail.Content,
// 		Attachments:      batchEmail.Attachments,
// 		Headers:          batchEmail.Headers,
// 		Sections:         batchEmail.Sections,
// 		Categories:       batchEmail.Categories,
// 		Personalizations: []Personalization{batchEmail.Personalization},
// 		Credentials:      credentials,
// 	}

// 	// Select sender based on adjusted weights
// 	sender := SelectSender(adjustedWeights)

// 	// Send email
// 	switch sender {
// 	case "sendgrid":
// 		log.Printf("Sending Message with %v", sender)
// 		SendEmailWithSendGrid(emailMessage)
// 	case "socketlabs":
// 		log.Printf("Sending Message with %v", sender)
// 		SendEmailWithSocketLabs(emailMessage)
// 	case "postmark":
// 		log.Printf("Sending Message with %v", sender)
// 		SendEmailWithPostmark(emailMessage)
// 	case "sparkpost":
// 		log.Printf("Sending Message with %v", sender)
// 		SendEmailWithSparkPost(emailMessage)
// 	default:
// 		log.Printf("No valid credentials found for sender: %s", sender)
// 	}

// 	// Update batch status
// 	err = updateBatchStatus(db, batchEmail.BatchID)
// 	if err != nil {
// 		log.Printf("Failed to update batch status: %v", err)
// 	}

// 	// Optionally, you could update the weights in the database here if you want to persist the adjusted weights
// 	// updateBatchWeights(db, batchEmail.BatchID, adjustedWeights)
// }

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

	batchInfo, err := createBatchRecord(db, userID, emailMessage, initialWeights)
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

func getBatchSize(emailMessage EmailMessage) int {
	var batchSize int
	if batchSizeValue, ok := emailMessage.CustomArgs["BatchSize"]; ok {
		switch v := batchSizeValue.(type) {
		case int:
			batchSize = v
		case float64:
			batchSize = int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				batchSize = i
			}
		default:
			// Handle unexpected type or set a default value
			batchSize = 0 // or some default value
		}
	}

	return batchSize

}

func getIntervalSeconds(emailMessage EmailMessage) int {
	var batchInterval int
	if batchIntervalValue, ok := emailMessage.CustomArgs["BatchInterval"]; ok {
		switch v := batchIntervalValue.(type) {
		case int:
			batchInterval = v
		case float64:
			batchInterval = int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				batchInterval = i
			}
		default:
			// Handle unexpected type or set a default value
			batchInterval = 0 // or some default value
		}
	}

	return batchInterval
}

func createBatchRecord(db *sql.DB, userID int, emailMessage EmailMessage, initialWeights map[string]int) (*BatchInfo, error) {
	intervalSeconds := getIntervalSeconds(emailMessage)
	batchSize := getBatchSize(emailMessage)

	batchInfo := &BatchInfo{
		UserID:          userID,
		TotalEmails:     len(emailMessage.Personalizations),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
		Status:          "pending",
		InitialWeights:  initialWeights,
		BatchSize:       batchSize,
		IntervalSeconds: intervalSeconds,
	}

	weightsJSON, err := json.Marshal(initialWeights)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal initial weights: %v", err)
	}

	err = db.QueryRow(`
        INSERT INTO email_batches (user_id, total_messages, created_at, updated_at, status, initial_weights, interval_seconds, batch_size)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `, batchInfo.UserID, batchInfo.TotalEmails, batchInfo.CreatedAt, batchInfo.UpdatedAt, batchInfo.Status, weightsJSON, batchInfo.IntervalSeconds, batchInfo.BatchSize).Scan(&batchInfo.ID)

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
	var totalMessages, processedEmails int
	var currentStatus string
	err = tx.QueryRow(`
        SELECT total_messages, processed_emails, status 
        FROM email_batches 
        WHERE id = $1 
        FOR UPDATE
    `, batchID).Scan(&totalMessages, &processedEmails, &currentStatus)
	if err != nil {
		return fmt.Errorf("failed to get batch info: %v", err)
	}

	// Increment the processed emails count
	processedEmails++

	// Determine the new status
	var newStatus string
	if processedEmails >= totalMessages {
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

func fetchBatchesDueForProcessing(db *sql.DB) ([]BatchInfo, error) {
	rows, err := db.Query(`
        SELECT id, batch_size, interval_seconds 
        FROM email_batches 
        WHERE status != 'completed' AND updated_at + (interval_seconds * INTERVAL '1 second') <= NOW() AT TIME ZONE 'UTC';
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []BatchInfo
	for rows.Next() {
		var batch BatchInfo
		if err := rows.Scan(&batch.ID, &batch.BatchSize, &batch.IntervalSeconds); err != nil {
			return nil, err
		}
		batches = append(batches, batch)
	}
	return batches, nil
}

// func processBatchMessages(consumer sarama.ConsumerGroup, batchID int) {
// 	ctx := context.Background()
// 	for {
// 		topics := []string{"batch_emails"}
// 		handler := &ConsumerGroupHandler{BatchID: batchID}
// 		err := consumer.Consume(ctx, topics, handler)
// 		if err != nil {
// 			log.Printf("Error from consumer: %v", err)
// 		}
// 		if ctx.Err() != nil {
// 			return
// 		}
// 	}
// }

func ManageBatchProcessing(kafkaBrokers []string, config *sarama.Config) {
	log.Println("Starting ManageBatchProcessing")

	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(kafkaBrokers, "batch-processor-group", config)
	if err != nil {
		log.Fatalf("Error creating consumer group client: %v", err)
	}
	defer consumer.Close()

	database.InitDB()
	db := database.GetDB()

	// Create a thread-safe map to store batch information
	var batchInfo sync.Map

	// Start the Kafka consumer
	go runKafkaConsumer(consumer, db, &batchInfo)

	// Start the topic message checker
	go func() {
		for {
			checkTopicMessages(kafkaBrokers, "batch_emails")
			time.Sleep(10 * time.Second)
		}
	}()

	// Start a goroutine to continuously check for batches that need processing
	go func() {
		for {
			batches, err := fetchBatchesDueForProcessing(db)
			if err != nil {
				log.Printf("Error fetching batches: %v", err)
				time.Sleep(10 * time.Second)
				continue
			}

			log.Printf("Fetched %d batches due for processing", len(batches))

			for _, batch := range batches {
				batchInfo.Store(batch.ID, batch)
				log.Printf("Added batch %d to batchInfo", batch.ID)
			}

			// Clean up completed batches
			batchInfo.Range(func(key, value interface{}) bool {
				id := key.(int)
				if isBatchCompleted(db, id) {
					batchInfo.Delete(id)
					log.Printf("Removed completed batch %d from batchInfo", id)
				}
				return true
			})

			time.Sleep(10 * time.Second)
		}
	}()

	log.Println("ManageBatchProcessing setup complete, waiting indefinitely")
	// Wait forever
	select {}
}

func runKafkaConsumer(consumer sarama.ConsumerGroup, db *sql.DB, batchInfo *sync.Map) {
	for {
		log.Println("Starting a new consumer session")
		handler := &ConsumerGroupHandler{
			db:        db,
			batchInfo: batchInfo,
		}
		err := consumer.Consume(context.Background(), []string{"batch_emails"}, handler)
		if err != nil {
			log.Printf("Error from consumer: %v", err)
		}
		log.Println("Consumer session ended")
	}
}

type ConsumerGroupHandler struct {
	db        *sql.DB
	batchInfo *sync.Map
}

func (h *ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("Starting to consume messages for partition %d, initial offset: %d",
		claim.Partition(), claim.InitialOffset())

	processedBatches := make(map[int]time.Time)

	for msg := range claim.Messages() {
		log.Printf("Received message: topic=%s partition=%d offset=%d", msg.Topic, msg.Partition, msg.Offset)

		batchID, err := strconv.Atoi(string(msg.Key))
		if err != nil {
			log.Printf("Invalid batch ID: %s", string(msg.Key))
			sess.MarkMessage(msg, "")
			continue
		}

		batchValue, exists := h.batchInfo.Load(batchID)
		if !exists {
			log.Printf("Batch %d not found in batchInfo", batchID)
			sess.MarkMessage(msg, "")
			continue
		}
		batch := batchValue.(BatchInfo)

		lastProcessed, exists := processedBatches[batchID]
		if exists && time.Since(lastProcessed) < time.Duration(batch.IntervalSeconds)*time.Second {
			log.Printf("Batch %d is not due for processing yet. Last processed: %v, Interval: %d seconds. Skipping.",
				batchID, lastProcessed, batch.IntervalSeconds)
			sess.MarkMessage(msg, "")
			continue
		}

		if !isBatchDueForProcessing(h.db, batch) {
			log.Printf("Batch %d is not due for processing yet based on database. Skipping.", batchID)
			sess.MarkMessage(msg, "")
			continue
		}

		log.Printf("Processing batch %d", batchID)

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

		err = json.Unmarshal(msg.Value, &batchEmail)
		if err != nil {
			log.Printf("Failed to parse batch email JSON: %v", err)
			sess.MarkMessage(msg, "")
			continue
		}

		// Fetch batch info including created_at and initial_weights
		var batchInfo struct {
			CreatedAt      time.Time
			InitialWeights string
		}
		err = h.db.QueryRow("SELECT created_at, initial_weights FROM email_batches WHERE id = $1", batchEmail.BatchID).Scan(
			&batchInfo.CreatedAt, &batchInfo.InitialWeights)
		if err != nil {
			log.Printf("Failed to fetch batch info: %v", err)
			sess.MarkMessage(msg, "")
			continue
		}

		var initialWeights map[string]int
		err = json.Unmarshal([]byte(batchInfo.InitialWeights), &initialWeights)
		if err != nil {
			log.Printf("Failed to unmarshal initial weights: %v", err)
			sess.MarkMessage(msg, "")
			continue
		}

		// Fetch ESP credentials
		credentials, err := fetchESPCredentialsForBatch(batchEmail.BatchID)
		if err != nil {
			log.Printf("Failed to fetch ESP credentials: %v", err)
			sess.MarkMessage(msg, "")
			continue
		}

		// Calculate new weights based on recent events
		// createdAtUnix := batchInfo.CreatedAt.Unix()
		newWeights, err := calculateRecentWeights(h.db, batchEmail.BatchID, batch.CreatedAt)
		if err != nil {
			log.Printf("Failed to calculate recent weights for BatchID: %v - CreatedAt: %v (Unix: %v), With Error: %v",
				batchEmail.BatchID, batchInfo.CreatedAt, batch.CreatedAt, err)
			// Fall back to initial weights if there's an error
			newWeights = initialWeights
		}

		// Compare new weights with initial weights and adjust
		adjustedWeights := adjustWeights(initialWeights, newWeights)

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
			log.Printf("Sending Message with %v", sender)
			SendEmailWithSendGrid(emailMessage)
		case "socketlabs":
			log.Printf("Sending Message with %v", sender)
			SendEmailWithSocketLabs(emailMessage)
		case "postmark":
			log.Printf("Sending Message with %v", sender)
			SendEmailWithPostmark(emailMessage)
		case "sparkpost":
			log.Printf("Sending Message with %v", sender)
			SendEmailWithSparkPost(emailMessage)
		default:
			log.Printf("No valid credentials found for sender: %s", sender)
			// fmt.Errorf("no valid credentials for sender: %s", sender)
		}

		// if sendErr != nil {
		// 	log.Printf("Failed to send email: %v", sendErr)
		// 	// Here you might want to implement some retry logic or error handling
		// } else {
		// 	log.Printf("Successfully sent email for batch %d", batchID)
		// }

		// Update batch status
		err = updateBatchStatus(h.db, batchID)
		if err != nil {
			log.Printf("Error updating batch status: %v", err)
		}

		processedBatches[batchID] = time.Now()
		sess.MarkMessage(msg, "")
	}

	log.Printf("Finished consuming messages for partition %d", claim.Partition())
	return nil
}

func isBatchDueForProcessing(db *sql.DB, batch BatchInfo) bool {
	var lastProcessed time.Time
	err := db.QueryRow("SELECT updated_at FROM email_batches WHERE id = $1", batch.ID).Scan(&lastProcessed)
	if err != nil {
		log.Printf("Error checking batch last processed time: %v", err)
		return false
	}
	timeSinceLastProcessed := time.Since(lastProcessed)
	isDue := timeSinceLastProcessed >= time.Duration(batch.IntervalSeconds)*time.Second
	log.Printf("Batch %d: Last processed %v ago, interval: %d seconds, is due: %v",
		batch.ID, timeSinceLastProcessed, batch.IntervalSeconds, isDue)
	return isDue
}

func isBatchCompleted(db *sql.DB, batchID int) bool {
	var status string
	err := db.QueryRow("SELECT status FROM email_batches WHERE id = $1", batchID).Scan(&status)
	if err != nil {
		log.Printf("Error checking batch status: %v", err)
		return false
	}
	return status == "completed"
}

func checkTopicMessages(brokers []string, topic string) {
	config := sarama.NewConfig()
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}
	defer client.Close()

	partitions, err := client.Partitions(topic)
	if err != nil {
		log.Printf("Error getting partitions: %v", err)
		return
	}

	for _, partition := range partitions {
		oldest, err := client.GetOffset(topic, partition, sarama.OffsetOldest)
		if err != nil {
			log.Printf("Error getting oldest offset: %v", err)
			continue
		}
		newest, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Printf("Error getting newest offset: %v", err)
			continue
		}
		log.Printf("Topic: %s, Partition: %d, Oldest Offset: %d, Newest Offset: %d, Messages: %d",
			topic, partition, oldest, newest, newest-oldest)
	}
}

// func fetchESPCredentialsForBatch(batchID int) (Credentials, error) {
// 	var cred Credentials
// 	database.InitDB()
// 	db := database.GetDB()

// 	query := `
//         SELECT esp.sendgrid_api_key, esp.sparkpost_api_key,
//                esp.socketlabs_api_key,
//                esp.socketlabs_server_id, esp.postmark_server_token
//         FROM public.email_service_providers esp
//         JOIN public.email_batches eb ON esp.user_id = eb.user_id
//         WHERE eb.id = $1
//     `
// 	err := db.QueryRow(query, batchID).Scan(
// 		&cred.SendgridAPIKey,
// 		&cred.SparkpostAPIKey,
// 		&cred.SocketLabsServerID,
// 		&cred.SocketLabsAPIKey,
// 		&cred.PostmarkServerToken,
// 	)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return cred, fmt.Errorf("no ESP credentials found for batch ID %d", batchID)
// 		}
// 		return cred, fmt.Errorf("error fetching ESP credentials: %v", err)
// 	}

// 	return cred, nil
// }

func fetchESPCredentialsForBatch(batchID int) (Credentials, error) {
	database.InitDB()
	db := database.GetDB()

	query := `
        SELECT esp.provider_name, esp.socketlabs_server_id, esp.socketlabs_api_key,
               esp.postmark_server_token,
               esp.sendgrid_api_key, 
			   esp.sparkpost_api_key
        FROM public.email_service_providers esp
        JOIN public.email_batches eb ON esp.user_id = eb.user_id
        WHERE eb.id = $1
    `
	rows, err := db.Query(query, batchID)
	if err != nil {
		return Credentials{}, fmt.Errorf("failed to query ESP credentials: %v", err)
	}
	defer rows.Close()

	creds := Credentials{}
	rowCount := 0

	for rows.Next() {
		rowCount++
		var (
			providerName,
			socketlabsServerID, socketlabsAPIKey,
			postmarkServerKey, sendgridAPIKey,
			sparkpostAPIKey sql.NullString
		)

		err := rows.Scan(
			&providerName,
			&socketlabsServerID, &socketlabsAPIKey,
			&postmarkServerKey,
			&sendgridAPIKey,
			&sparkpostAPIKey,
		)
		if err != nil {
			return Credentials{}, fmt.Errorf("failed to scan ESP credential: %v", err)
		}

		switch providerName.String {
		case "socketlabs":
			creds.SocketLabsServerID = socketlabsServerID.String
			creds.SocketLabsAPIKey = socketlabsAPIKey.String
		case "postmark":
			creds.PostmarkServerToken = postmarkServerKey.String
		case "sendgrid":
			creds.SendgridAPIKey = sendgridAPIKey.String
		case "sparkpost":
			creds.SparkpostAPIKey = sparkpostAPIKey.String
		default:
			log.Printf("Unknown provider: %s", providerName.String)
		}
	}

	if err := rows.Err(); err != nil {
		return Credentials{}, fmt.Errorf("error iterating over rows: %v", err)
	}

	if rowCount == 0 {
		return Credentials{}, fmt.Errorf("no ESP credentials found for user ID %d", batchID)
	}

	return creds, nil
}
