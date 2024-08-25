package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/IBM/sarama"
)

func ProcessEmailMessages(msg *sarama.ConsumerMessage) {
	serverID, _ := strconv.Atoi(os.Getenv("SOCKETLABS_SERVER_ID"))
	apiKey := os.Getenv("SOCKETLABS_API_KEY")

	var emailMessage EmailMessage
	err := json.Unmarshal(msg.Value, &emailMessage)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	for _, recipient := range emailMessage.To {
		// Create a new EmailMessage for each recipient
		individualEmail := EmailMessage{
			From:           emailMessage.From,
			To:             []EmailAddress{recipient}, // Single recipient
			Cc:             emailMessage.Cc,
			Bcc:            emailMessage.Bcc,
			Subject:        emailMessage.Subject,
			Body:           emailMessage.Body,
			Attachments:    emailMessage.Attachments,
			Headers:        emailMessage.Headers,
			AdditionalData: emailMessage.AdditionalData,
		}

		// Select sender based on weights
		sender := SelectSender()
		if sender == "SocketLabs" {
			SendEmailWithSocketLabs(serverID, apiKey, individualEmail)
		} else {
			SendEmailWithPostmark(individualEmail)
		}
	}
}

func SelectSender() string {
	socketLabsWeight, err := strconv.Atoi(os.Getenv("SOCKETLABS_WEIGHT"))
	if err != nil {
		log.Printf("Invalid SOCKETLABS_WEIGHT, using default 70")
		socketLabsWeight = 70
	}
	postmarkWeight, err := strconv.Atoi(os.Getenv("POSTMARK_WEIGHT"))
	if err != nil {
		log.Printf("Invalid POSTMARK_WEIGHT, using default 30")
		postmarkWeight = 30
	}
	totalWeight := socketLabsWeight + postmarkWeight
	randomValue := rand.Intn(totalWeight)
	if randomValue < socketLabsWeight {
		return "SocketLabs"
	}
	return "Postmark"
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
	Body           string            `json:"body"`
	Attachments    []Attachment      `json:"attachments"`
	Headers        map[string]string `json:"headers"`
	AdditionalData map[string]string `json:"additionaldata"`
}
