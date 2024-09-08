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

	preparedMessages := prepareSocketLabsMessages(emailMessage)

	// Print prepared messages for review
	printPreparedMessages(preparedMessages)

	// Optional: Add a prompt to continue or abort
	fmt.Print("Press Enter to continue sending, or Ctrl+C to abort...")
	// fmt.Scanln() // Wait for user input

	for _, basic := range preparedMessages {
		res, err := client.SendBasic(basic)
		if err != nil {
			errorHandler.HandleSendError(basic.To[0].EmailAddress, err, &res)
			return fmt.Errorf("failed to send email: %v", err)
		}
		if res.Result != injectionapi.SendResultSUCCESS {
			err := fmt.Errorf("send unsuccessful: %s - %s", res.Result.ToString(), res.Result.ToResponseMessage())
			errorHandler.HandleSendError(basic.To[0].EmailAddress, err, &res)
			return err
		}
	}

	return nil
}

func prepareSocketLabsMessages(emailMessage EmailMessage) []*message.BasicMessage {
	xxsMessageId := generateXxsMessageId(emailMessage.Credentials.SocketLabsAPIKey)
	parsedSections := parseSectionsDynamic(emailMessage.Sections)
	var preparedMessages []*message.BasicMessage

	for _, personalization := range emailMessage.Personalizations {
		processedContent := processContent(emailMessage.Content, personalization.Substitutions, parsedSections)

		basic := &message.BasicMessage{
			Subject: personalization.Subject,
			From: message.EmailAddress{
				EmailAddress: emailMessage.From.Email,
				FriendlyName: emailMessage.From.Name,
			},
			PlainTextBody: getContentByType(processedContent, "text/plain"),
			HtmlBody:      getContentByType(processedContent, "text/html"),
		}

		basic.AddToEmailAddress(personalization.To.Email)

		for _, cc := range emailMessage.Cc {
			basic.AddCcEmailAddress(cc)
		}
		for _, bcc := range emailMessage.Bcc {
			basic.AddBccEmailAddress(bcc)
		}

		for _, attachment := range emailMessage.Attachments {
			content, err := base64.StdEncoding.DecodeString(attachment.Content)
			if err != nil {
				log.Printf("Failed to decode attachment content: %v", err)
				continue
			}
			basic.Attachments = append(basic.Attachments, message.Attachment{
				Content:  content,
				MimeType: attachment.Type,
				Name:     attachment.Name,
			})
		}

		basic.CustomHeaders = append(basic.CustomHeaders, message.CustomHeader{Name: "X-xsMessageId", Value: xxsMessageId})
		for key, value := range emailMessage.Headers {
			basic.CustomHeaders = append(basic.CustomHeaders, message.CustomHeader{Name: key, Value: value})
		}

		preparedMessages = append(preparedMessages, basic)
	}

	return preparedMessages
}

func generateXxsMessageId(apiKey string) string {
	// Create a unique string using apiKey and current timestamp
	uniqueString := fmt.Sprintf("%s-%d", apiKey, time.Now().UnixNano())

	// Create MD5 hash
	hasher := md5.New()
	hasher.Write([]byte(uniqueString))
	return hex.EncodeToString(hasher.Sum(nil))
}

func printPreparedMessages(preparedMessages []*message.BasicMessage) {
	for i, msg := range preparedMessages {
		fmt.Printf("Message %d:\n", i+1)
		fmt.Printf("From: %s (%s)\n", msg.From.EmailAddress, msg.From.FriendlyName)
		fmt.Printf("To: %v\n", msg.To)
		fmt.Printf("Subject: %s\n", msg.Subject)
		fmt.Printf("Plain Text Body: %s\n", msg.PlainTextBody)
		fmt.Printf("HTML Body: %s\n", msg.HtmlBody)
		fmt.Println("Custom Headers:")
		for _, header := range msg.CustomHeaders {
			fmt.Printf(" - %s: %s\n", header.Name, header.Value)
		}
		fmt.Println("Attachments:")
		for _, att := range msg.Attachments {
			fmt.Printf(" - %s (%s)\n", att.Name, att.MimeType)
		}
	}
}
