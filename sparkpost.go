package main

import (
	"encoding/base64"
	"log"

	sp "github.com/SparkPost/gosparkpost"
)

func SendEmailWithSparkPost(emailMessage EmailMessage) {
	// Get the API key from the credentials
	apiKey := emailMessage.Credentials.SparkpostAPIKey
	if apiKey == "" {
		log.Fatal("Missing SparkPost API key in credentials")
	}

	// Configure SparkPost client
	cfg := &sp.Config{
		BaseUrl:    "https://api.sparkpost.com",
		ApiKey:     apiKey,
		ApiVersion: 1,
	}
	var client sp.Client
	err := client.Init(cfg)
	if err != nil {
		log.Fatalf("SparkPost client init failed: %s\n", err)
	}

	// Prepare recipients
	recipients := make([]sp.Recipient, len(emailMessage.To))
	for i, addr := range emailMessage.To {
		recipients[i] = sp.Recipient{
			Address: sp.Address{
				Email: addr.Email,
			},
		}
	}

	// Prepare attachments
	attachments := make([]sp.Attachment, len(emailMessage.Attachments))
	for i, att := range emailMessage.Attachments {
		content, err := base64.StdEncoding.DecodeString(att.Content)
		if err != nil {
			log.Printf("Failed to decode attachment content: %v", err)
			continue
		}
		attachments[i] = sp.Attachment{
			Filename: att.Name,
			MIMEType: att.ContentType,
			B64Data:  string(content),
		}
	}

	// Create a Transmission
	tx := &sp.Transmission{
		Recipients: recipients,
		Content: sp.Content{
			From:        sp.Address{Email: emailMessage.From.Email, Name: emailMessage.From.Name},
			Subject:     emailMessage.Subject,
			HTML:        emailMessage.HtmlBody,
			Text:        emailMessage.TextBody,
			Headers:     emailMessage.Headers,
			Attachments: attachments,
		},
	}

	// Send the email
	id, _, err := client.Send(tx)
	if err != nil {
		log.Fatalf("Failed to send email with SparkPost: %v", err)
	}

	log.Printf("Email sent with SparkPost. Transmission ID: %s", id)
}
