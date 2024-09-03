package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/IBM/sarama"
)

type KafkaMessage struct {
	Headers map[string][]string `json:"headers"`
	Body    EmailMessage        `json:"body"`
}

func ProcessEmailMessages(msg *sarama.ConsumerMessage) {
	var kafkaMessage KafkaMessage
	err := json.Unmarshal(msg.Value, &kafkaMessage)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	emailMessage := kafkaMessage.Body
	// Extract credentials from the email message
	credentials := emailMessage.Credentials
	socketLabsWeight, postmarkWeight, sendGridWeight, sparkPostWeight := calculateWeights(credentials)

	// Select sender based on available weights

	for _, recipient := range emailMessage.To {
		// Create a new EmailMessage for each recipient
		individualEmail := EmailMessage{
			From:           emailMessage.From,
			To:             []EmailAddress{recipient}, // Single recipient
			Cc:             emailMessage.Cc,
			Bcc:            emailMessage.Bcc,
			Subject:        emailMessage.Subject,
			TextBody:       emailMessage.TextBody,
			HtmlBody:       emailMessage.HtmlBody,
			Attachments:    emailMessage.Attachments,
			Headers:        emailMessage.Headers,
			AdditionalData: emailMessage.AdditionalData,
			Credentials:    emailMessage.Credentials,
		}

		sender := SelectSender(socketLabsWeight, postmarkWeight, sendGridWeight, sparkPostWeight)
		switch sender {
		case "SocketLabs":
			SendEmailWithSocketLabs(individualEmail)
		case "Postmark":
			SendEmailWithPostmark(individualEmail)
		case "SendGrid":
			SendEmailWithSendGrid(individualEmail)
		case "SparkPost":
			SendEmailWithSparkPost(individualEmail)
		default:
			log.Printf("No valid credentials found for any sender for recipient: %s", recipient.Email)
		}
	}
}

func SelectSender(socketLabsWeight, postmarkWeight, sendGridWeight, sparkPostWeight int) string {
	totalWeight := socketLabsWeight + postmarkWeight + sendGridWeight + sparkPostWeight
	if totalWeight == 0 {
		return ""
	}

	randomValue := rand.Intn(totalWeight)
	if randomValue < socketLabsWeight {
		return "SocketLabs"
	} else if randomValue < socketLabsWeight+postmarkWeight {
		return "Postmark"
	} else if randomValue < socketLabsWeight+postmarkWeight+sendGridWeight {
		return "SendGrid"
	}
	return "SparkPost"
}

func calculateWeights(credentials Credentials) (int, int, int, int) {
	// Parse weights, default to 0 if not present
	socketLabsWeight, _ := strconv.Atoi(credentials.SocketLabsWeight)
	postmarkWeight, _ := strconv.Atoi(credentials.PostmarkWeight)
	sendGridWeight, _ := strconv.Atoi(credentials.SendgridWeight)
	sparkPostWeight, _ := strconv.Atoi(credentials.SparkpostWeight)

	// Check for available credentials and adjust weights accordingly
	if credentials.SocketLabsServerID == "" || credentials.SocketLabsAPIKey == "" {
		socketLabsWeight = 0
	}
	if credentials.PostmarkServerToken == "" {
		postmarkWeight = 0
	}
	if credentials.SendgridAPIKey == "" {
		sendGridWeight = 0
	}
	if credentials.SparkpostAPIKey == "" {
		sparkPostWeight = 0
	}

	return socketLabsWeight, postmarkWeight, sendGridWeight, sparkPostWeight
}

type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name"`
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
	Name        string `json:"Name"`
	ContentType string `json:"ContentType"`
	Content     string `json:"Content"`
}

type EmailMessage struct {
	From           EmailAddress      `json:"from"`
	To             []EmailAddress    `json:"to"`
	Cc             []string          `json:"cc"`
	Bcc            []string          `json:"bcc"`
	Subject        string            `json:"subject"`
	TextBody       string            `json:"textbody"`
	HtmlBody       string            `json:"htmlbody"`
	Attachments    []Attachment      `json:"attachments"`
	Headers        map[string]string `json:"headers"`
	AdditionalData map[string]string `json:"additionaldata"`
	Credentials    Credentials       `json:"credentials"`
}

type Credentials struct {
	SocketLabsServerID  string `json:"SocketLabsServerID"`
	SocketLabsAPIKey    string `json:"SocketLabsAPIkey"`
	SocketLabsWeight    string `json:"SocketLabsWeight"`
	PostmarkServerToken string `json:"PostmarkServerToken"`
	// PostmarkAPIURL      string `json:"PostmarkAPIURL"`
	PostmarkWeight  string `json:"PostmarkWeight"`
	SendgridAPIKey  string `json:"SendgridAPIKey"`
	SendgridWeight  string `json:"SendgridWeight"`
	SparkpostAPIKey string `json:"SparkpostAPIKey"`
	SparkpostWeight string `json:"SparkpostWeight"`
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
