package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/socketlabs/socketlabs-go/injectionapi"
	"github.com/socketlabs/socketlabs-go/injectionapi/message"
)

func SendEmailWithSocketLabs(emailMessage EmailMessage) error {
	// Use credentials from the EmailMessage
	serverID, _ := strconv.Atoi(emailMessage.Credentials.SocketLabsServerID)
	apiKey := emailMessage.Credentials.SocketLabsAPIKey

	client := injectionapi.CreateClient(serverID, apiKey)
	errorHandler := NewSocketLabsErrorHandler()

	// Generate X-xsMessageId
	xxsMessageId := generateXxsMessageId(apiKey)

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
				MimeType: attachment.Type,
				Name:     attachment.Name,
			}
			basic.Attachments = append(basic.Attachments, socketLabsAttachment)
		}

		// Initialize a map to store all headers
		allHeaders := make(map[string]string)

		// Add the custom X-xsMessageId header
		allHeaders["X-xsMessageId"] = xxsMessageId

		// Add all headers from emailMessage.Headers
		for key, value := range emailMessage.Headers {
			allHeaders[key] = value
		}

		// Clear existing CustomHeaders to avoid duplication
		basic.CustomHeaders = nil

		// Add all headers to CustomHeaders
		for key, value := range allHeaders {
			basic.CustomHeaders = append(basic.CustomHeaders, message.CustomHeader{Name: key, Value: value})
		}

		log.Printf("%v", &basic)
		// Send the email
		res, err := client.SendBasic(&basic)

		if err != nil {
			errorHandler.HandleSendError(to.Email, err, &res)
			return fmt.Errorf("failed to send email: %v", err)
		}
		if res.Result != injectionapi.SendResultSUCCESS {
			err := fmt.Errorf("send unsuccessful: %s - %s", res.Result.ToString(), res.Result.ToResponseMessage())
			errorHandler.HandleSendError(to.Email, err, &res)
			return err
		}

	}

	return nil
}

func generateXxsMessageId(apiKey string) string {
	// Create a unique string using apiKey and current timestamp
	uniqueString := fmt.Sprintf("%s-%d", apiKey, time.Now().UnixNano())

	// Create MD5 hash
	hasher := md5.New()
	hasher.Write([]byte(uniqueString))
	return hex.EncodeToString(hasher.Sum(nil))
}
