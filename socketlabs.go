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
	serverID, _ := strconv.Atoi(emailMessage.Credentials.SocketLabsServerID)
	apiKey := emailMessage.Credentials.SocketLabsAPIKey

	client := injectionapi.CreateClient(serverID, apiKey)
	errorHandler := NewSocketLabsErrorHandler()

	xxsMessageId := generateXxsMessageId(apiKey)

	// Parse sections
	parsedSections := parseSectionsDynamic(emailMessage.Sections)

	// Process personalizations
	for _, personalization := range emailMessage.Personalizations {
		// Process content for each personalization
		processedContent := processContent(emailMessage.Content, personalization.Substitutions, parsedSections)

		basic := message.BasicMessage{
			Subject: personalization.Subject,
			From: message.EmailAddress{
				EmailAddress: emailMessage.From.Email,
				FriendlyName: emailMessage.From.Name,
			},
			PlainTextBody: getContentByType(processedContent, "text/plain"),
			HtmlBody:      getContentByType(processedContent, "text/html"),
		}

		// Add the recipient
		basic.AddToEmailAddress(personalization.To.Email)

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

		// Add headers
		allHeaders := make(map[string]string)
		allHeaders["X-xsMessageId"] = xxsMessageId
		for key, value := range emailMessage.Headers {
			allHeaders[key] = value
		}
		basic.CustomHeaders = nil
		for key, value := range allHeaders {
			basic.CustomHeaders = append(basic.CustomHeaders, message.CustomHeader{Name: key, Value: value})
		}

		// Send the email
		res, err := client.SendBasic(&basic)
		if err != nil {
			errorHandler.HandleSendError(personalization.To.Email, err, &res)
			return fmt.Errorf("failed to send email: %v", err)
		}
		if res.Result != injectionapi.SendResultSUCCESS {
			err := fmt.Errorf("send unsuccessful: %s - %s", res.Result.ToString(), res.Result.ToResponseMessage())
			errorHandler.HandleSendError(personalization.To.Email, err, &res)
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
