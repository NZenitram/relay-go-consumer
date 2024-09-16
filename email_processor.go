package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"relay-go-consumer/database"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

type KafkaMessage struct {
	BatchID   int
	MessageID string              `json:"MessageID"`
	UserID    int                 `json:"UserID"`
	Headers   map[string][]string `json:"headers"`
	Body      EmailMessage        `json:"body"`
}

type BatchInfo struct {
	BatchID         int
	TotalMessages   int
	BatchSize       int
	IntervalSeconds int
	StartTime       time.Time
	EndTime         sql.NullTime
	CurrentBatch    int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	TotalBatches    sql.NullInt32
	UserID          sql.NullInt32
	BatchesToKafka  sql.NullInt32
	Status          sql.NullString
}

func ProcessEmailMessages(msg *sarama.ConsumerMessage) {
	var kafkaMessage KafkaMessage
	err := json.Unmarshal(msg.Value, &kafkaMessage)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	batchID := kafkaMessage.BatchID

	database.InitDB()
	db := database.GetDB()

	// Fetch ESP credentials from the database
	credentials, credentialsErr := fetchESPCredentials(kafkaMessage.UserID)
	if credentialsErr != nil {
		fmt.Printf("failed to fetch ESP credentials: %v", credentialsErr)
		return
	}

	emailMessage := kafkaMessage.Body
	emailMessage.Credentials = credentials

	// Calculate weights based on event data
	var weights map[string]int
	if batchID != 0 {
		batchInfo, err := fetchBatchData(db, batchID)
		if err != nil {
			fmt.Printf("Failed to fetch batch data: %v", err)
			return
		}

		currentTime := time.Now().UTC()
		var startTime, endTime time.Time

		if batchInfo.CurrentBatch < 1 {
			// For the first batch, use the last 30 days
			startTime = currentTime.AddDate(0, 0, -30)
			endTime = currentTime
		} else {
			// For subsequent batches, use the time since the last batch
			startTime = batchInfo.UpdatedAt
			endTime = currentTime.Add(time.Duration(batchInfo.IntervalSeconds) * time.Second)
		}

		weights, err = calculateWeightsForTimeRange(db, kafkaMessage.UserID, credentials, startTime, endTime)
		if err != nil {
			fmt.Printf("failed to calculate weights: %v", err)
			return
		}
	} else {
		// If batchID is 0, use the default 30 days back
		endTime := time.Now()
		startTime := endTime.AddDate(0, 0, -30)
		weights, err = calculateWeightsForTimeRange(db, kafkaMessage.UserID, credentials, startTime, endTime)
		if err != nil {
			fmt.Printf("failed to calculate weights: %v", err)
			return
		}
	}

	sendEmailsImmediately(emailMessage, weights)
}

func sendEmailsImmediately(emailMessage EmailMessage, weights map[string]int) {
	// If there are no personalizations, create one for each recipient
	if len(emailMessage.Personalizations) == 0 {
		for _, recipient := range emailMessage.To {
			emailMessage.Personalizations = append(emailMessage.Personalizations, Personalization{
				To:            recipient,
				Subject:       emailMessage.Subject,
				Substitutions: make(map[string]string),
			})
		}
	}

	senderGroups := make(map[string][]Personalization)
	for _, p := range emailMessage.Personalizations {
		sender := SelectSender(weights)
		senderGroups[sender] = append(senderGroups[sender], p)
	}
	// Send emails using each selected sender
	for sender, personalizations := range senderGroups {
		groupMessage := emailMessage
		groupMessage.Personalizations = personalizations
		switch sender {
		case "sendgrid":
			SendEmailWithSendGrid(groupMessage)
		case "socketlabs":
			SendEmailWithSocketLabs(groupMessage)
		case "postmark":
			SendEmailWithPostmark(groupMessage)
		case "sparkpost":
			SendEmailWithSparkPost(groupMessage)
		default:
			log.Printf("No valid credentials found for sender: %s", sender)
		}
	}
}

func SelectSender(weights map[string]int) string {
	totalWeight := 0
	for _, weight := range weights {
		totalWeight += weight
	}

	if totalWeight == 0 {
		return ""
	}

	randomValue := rand.Intn(totalWeight)
	cumulativeWeight := 0

	for provider, weight := range weights {
		cumulativeWeight += weight
		if randomValue < cumulativeWeight {
			return provider
		}
	}

	return "" // This should never happen if weights are calculated correctly
}

// Custom unmarshaling logic for EmailAddress
func (e *EmailAddress) UnmarshalJSON(data []byte) error {
	// Attempt to unmarshal as a simple string
	var emailString string
	if err := json.Unmarshal(data, &emailString); err == nil {
		// Updated regex pattern to capture Friendly Name and Email Address
		r := regexp.MustCompile(`(?i)(?:"?([^"<]*)"?\s*<([^>]+)>|([^<>\s]+@[^<>\s]+))`)
		matches := r.FindStringSubmatch(emailString)
		if len(matches) > 0 {
			e.Name = strings.TrimSpace(matches[1])
			if matches[2] != "" {
				e.Email = strings.TrimSpace(matches[2])
			} else {
				e.Email = strings.TrimSpace(matches[3])
			}
		} else {
			e.Email = emailString
			e.Name = "" // No name available
		}
		return nil
	}

	// Attempt to unmarshal as an object with email and name
	var alias struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	e.Email = alias.Email
	e.Name = alias.Name
	return nil
}

type Attachment struct {
	Content     string `json:"content"`
	ContentID   string `json:"content_id"`
	Disposition string `json:"disposition"`
	Filename    string `json:"filename"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	ContentType string `json:"ContentType,omitempty"`
}

type EmailMessage struct {
	From             EmailAddress
	To               []EmailAddress
	Cc               []string
	Bcc              []string
	Subject          string
	TextBody         string
	HtmlBody         string
	Content          []Content
	Attachments      []Attachment
	Headers          map[string]string
	CustomArgs       map[string]interface{} `json:"custom_args"`
	Credentials      Credentials
	Personalizations []Personalization
	Sections         map[string]string
	Categories       []string
}
type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Personalization struct {
	To            EmailAddress
	Subject       string
	Substitutions map[string]string
}

type EmailAddress struct {
	Name  string
	Email string
}

type Credentials struct {
	SocketLabsServerID  string `json:"SocketLabsServerID"`
	SocketLabsAPIKey    string `json:"SocketLabsAPIkey"`
	SocketLabsWeight    string `json:"SocketLabsWeight"`
	PostmarkServerToken string `json:"PostmarkServerToken"`
	PostmarkWeight      string `json:"PostmarkWeight"`
	SendgridAPIKey      string `json:"SendgridAPIKey"`
	SendgridWeight      string `json:"SendgridWeight"`
	SparkpostAPIKey     string `json:"SparkpostAPIKey"`
	SparkpostWeight     string `json:"SparkpostWeight"`
}

type StandardizedEvent struct {
	MessageID        string
	Provider         string
	Processed        bool
	ProcessedTime    int64
	Delivered        bool
	DeliveredTime    *int64
	Bounce           bool
	BounceType       string
	BounceTime       *int64
	Deferred         bool
	DeferredCount    int
	LastDeferralTime *int64
	UniqueOpen       bool
	UniqueOpenTime   *int64
	Open             bool
	OpenCount        int
	LastOpenTime     *int64
	Dropped          bool
	DroppedTime      *int64
	DroppedReason    string
}

type ESPCredential struct {
	ProviderName             string
	SendingDomains           []string
	SendgridVerificationKey  string
	SparkpostWebhookUser     string
	SparkpostWebhookPassword string
	SocketlabsSecretKey      string
	PostmarkWebhookUser      string
	PostmarkWebhookPassword  string
	SocketlabsServerID       string
}

func fetchESPCredentials(userID int) (Credentials, error) {
	database.InitDB()
	db := database.GetDB()

	query := `
        SELECT provider_name, 
               socketlabs_server_id, socketlabs_api_key,
               postmark_server_token,
               sendgrid_api_key,
               sparkpost_api_key,
			   weight
        FROM email_service_providers
        WHERE user_id = $1
    `
	rows, err := db.Query(query, userID)
	if err != nil {
		return Credentials{}, fmt.Errorf("failed to query ESP credentials: %v", err)
	}
	defer rows.Close()

	creds := Credentials{}
	rowCount := 0

	for rows.Next() {
		rowCount++
		var (
			providerName, socketlabsServerID, socketlabsAPIKey,
			postmarkServerKey, sendgridAPIKey,
			sparkpostAPIKey, senderWeight sql.NullString
		)

		err := rows.Scan(
			&providerName,
			&socketlabsServerID, &socketlabsAPIKey,
			&postmarkServerKey,
			&sendgridAPIKey,
			&sparkpostAPIKey,
			&senderWeight,
		)
		if err != nil {
			return Credentials{}, fmt.Errorf("failed to scan ESP credential: %v", err)
		}

		switch providerName.String {
		case "socketlabs":
			if socketlabsServerID.Valid && socketlabsAPIKey.Valid {
				creds.SocketLabsServerID = socketlabsServerID.String
				creds.SocketLabsAPIKey = socketlabsAPIKey.String
				creds.SocketLabsWeight = senderWeight.String
			}
		case "postmark":
			if postmarkServerKey.Valid {
				creds.PostmarkServerToken = postmarkServerKey.String
				creds.PostmarkWeight = senderWeight.String
			}
		case "sendgrid":
			if sendgridAPIKey.Valid {
				creds.SendgridAPIKey = sendgridAPIKey.String
				creds.SendgridWeight = senderWeight.String
			}
		case "sparkpost":
			if sparkpostAPIKey.Valid {
				creds.SparkpostAPIKey = sparkpostAPIKey.String
				creds.SparkpostWeight = senderWeight.String
			}
		default:
			log.Printf("Unknown provider: %s", providerName.String)
		}
	}

	if err := rows.Err(); err != nil {
		return Credentials{}, fmt.Errorf("error iterating over rows: %v", err)
	}

	if rowCount == 0 {
		return Credentials{}, fmt.Errorf("no ESP credentials found for user ID %d", userID)
	}

	return creds, nil
}
