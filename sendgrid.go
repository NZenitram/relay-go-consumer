package main

import (
	"encoding/base64"
	"log"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendEmailWithSendGrid(emailMessage EmailMessage) {
	// Use credentials from the EmailMessage
	apiKey := emailMessage.Credentials.SendgridAPIKey

	client := sendgrid.NewSendClient(apiKey)

	// Iterate over each recipient in the "To" field
	for _, to := range emailMessage.To {
		// Create the email message for each recipient
		from := mail.NewEmail(emailMessage.From.Name, emailMessage.From.Email)
		subject := emailMessage.Subject
		toEmail := mail.NewEmail(to.Name, to.Email)
		plainTextContent := emailMessage.TextBody
		htmlContent := emailMessage.HtmlBody
		message := mail.NewSingleEmail(from, subject, toEmail, plainTextContent, htmlContent)

		// Add CC and BCC recipients
		for _, cc := range emailMessage.Cc {
			ccEmail := mail.NewEmail("", cc)
			message.Personalizations[0].AddCCs(ccEmail)
		}

		for _, bcc := range emailMessage.Bcc {
			bccEmail := mail.NewEmail("", bcc)
			message.Personalizations[0].AddBCCs(bccEmail)
		}

		// Add attachments
		for _, attachment := range emailMessage.Attachments {
			content, err := base64.StdEncoding.DecodeString(attachment.Content)
			if err != nil {
				log.Printf("Failed to decode attachment content: %v", err)
				continue
			}
			sgAttachment := mail.NewAttachment()
			sgAttachment.SetContent(string(content))
			sgAttachment.SetType(attachment.ContentType)
			sgAttachment.SetFilename(attachment.Name)
			message.AddAttachment(sgAttachment)
		}

		// Add custom headers
		for key, value := range emailMessage.Headers {
			message.SetHeader(key, value)
		}

		// Send the email
		_, err := client.Send(message)
		if err != nil {
			log.Printf("Failed to send email to %s: %v", to.Email, err)
		}
	}
}
