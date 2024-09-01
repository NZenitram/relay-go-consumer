package main

import (
	"encoding/base64"
	"log"
	"strconv"

	"github.com/socketlabs/socketlabs-go/injectionapi"
	"github.com/socketlabs/socketlabs-go/injectionapi/message"
)

func SendEmailWithSocketLabs(emailMessage EmailMessage) {
	// Use credentials from the EmailMessage
	serverID, _ := strconv.Atoi(emailMessage.Credentials.SocketLabsServerID)
	apiKey := emailMessage.Credentials.SocketLabsAPIKey

	client := injectionapi.CreateClient(serverID, apiKey)

	// Iterate over each recipient in the "To" field
	for _, to := range emailMessage.To {
		// Create the email message for each recipient
		basic := message.BasicMessage{
			Subject: emailMessage.Subject,
			From: message.EmailAddress{
				EmailAddress: emailMessage.From.Email,
				FriendlyName: emailMessage.From.Name,
			},
			PlainTextBody: emailMessage.TextBody,
			HtmlBody:      emailMessage.HtmlBody,
		}

		// Add the recipient with a friendly name
		basic.AddToEmailAddress(to.Email)

		// Add CC and BCC recipients
		for _, cc := range emailMessage.Cc {
			basic.AddCcEmailAddress(cc)
		}

		for _, bcc := range emailMessage.Bcc {
			basic.AddBccEmailAddress(bcc)
		}

		// Add attachments
		for _, attachment := range emailMessage.Attachments {
			content, err := base64.StdEncoding.DecodeString(attachment.Content)
			if err != nil {
				log.Printf("Failed to decode attachment content: %v", err)
				continue
			}
			socketLabsAttachment := message.Attachment{
				Content:  content,
				MimeType: attachment.ContentType,
				Name:     attachment.Name,
			}
			basic.Attachments = append(basic.Attachments, socketLabsAttachment)
		}

		// Add custom headers
		for key, value := range emailMessage.Headers {
			basic.CustomHeaders = append(basic.CustomHeaders, message.CustomHeader{Name: key, Value: value})
		}

		// Send the email
		response, err := client.SendBasic(&basic)
		if err != nil {
			log.Printf("Failed to send email to %s: %v", to.Email, err)
		} else {
			log.Printf("Email sent to %s. Response: %v", to.Email, response)
		}
	}
}
