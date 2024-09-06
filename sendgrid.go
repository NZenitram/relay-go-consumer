package main

import (
	"encoding/base64"
	"log"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendEmailWithSendGrid(emailMessage EmailMessage) {
	apiKey := emailMessage.Credentials.SendgridAPIKey
	client := sendgrid.NewSendClient(apiKey)

	for _, p := range emailMessage.Personalizations {
		message := mail.NewV3Mail()

		from := mail.NewEmail(emailMessage.From.Name, emailMessage.From.Email)
		message.SetFrom(from)

		// Set subject (personalized or default)
		subject := p.Subject
		if subject == "" {
			subject = emailMessage.Subject
		}
		message.Subject = subject

		// Add recipient
		to := mail.NewEmail(p.To.Name, p.To.Email)
		personalization := mail.NewPersonalization()
		personalization.AddTos(to)

		// Add CC and BCC recipients
		for _, cc := range emailMessage.Cc {
			personalization.AddCCs(mail.NewEmail("", cc))
		}
		for _, bcc := range emailMessage.Bcc {
			personalization.AddBCCs(mail.NewEmail("", bcc))
		}

		// Process content with substitutions
		for _, content := range emailMessage.Content {
			processedValue := content.Value
			for key, value := range p.Substitutions {
				processedValue = strings.ReplaceAll(processedValue, key, value)
			}
			for key, value := range emailMessage.Sections {
				processedValue = strings.ReplaceAll(processedValue, key, value)
			}
			message.AddContent(mail.NewContent(content.Type, processedValue))
		}

		// Add substitutions
		personalization.Substitutions = p.Substitutions

		message.AddPersonalizations(personalization)

		// Add attachments
		for _, attachment := range emailMessage.Attachments {
			content, err := base64.StdEncoding.DecodeString(attachment.Content)
			if err != nil {
				log.Printf("Failed to decode attachment content: %v", err)
				continue
			}

			message.AddAttachment(&mail.Attachment{
				Content:     string(content),
				ContentID:   attachment.ContentID,
				Type:        attachment.Type,
				Filename:    attachment.Filename,
				Name:        attachment.Name,
				Disposition: attachment.Disposition,
			})
		}

		// Add custom headers
		message.Headers = emailMessage.Headers

		// // Add custom args
		// message.CustomArgs = emailMessage.CustomArgs

		// // Add categories
		message.Categories = emailMessage.Categories

		// Send the email
		response, err := client.Send(message)
		if err != nil {
			log.Printf("Error sending email to %s: %v", p.To.Email, err)
		} else {
			log.Printf("Email sent to %s. Status: %d, Body: %s", p.To.Email, response.StatusCode, response.Body)
		}
	}
}
