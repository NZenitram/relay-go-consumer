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

// // Map of tracking types to descriptions
// var trackingTypeMap = map[int]string{
// 	0: "Click",
// 	1: "Open",
// 	2: "Unsubscribe",
// }

// // Function to get the description for a tracking type
// func getTrackingTypeDescription(trackingType int) string {
// 	if description, exists := trackingTypeMap[trackingType]; exists {
// 		return description
// 	}
// 	return "Unknown"
// }

// Base struct for common fields
type SocketLabsBaseEvent struct {
	Provider     string `json:"Provider"`
	Type         string `json:"Type"`
	DateTime     string `json:"DateTime"`
	MailingId    string `json:"MailingId"`
	MessageId    string `json:"MessageId"`
	Address      string `json:"Address"`
	ServerId     int    `json:"ServerId"`
	SubaccountId int    `json:"SubaccountId"`
	IpPoolId     int    `json:"IpPoolId"`
	SecretKey    string `json:"SecretKey"`
}

// SocketLabsComplaintEvent struct
type SocketLabsComplaintEvent struct {
	SocketLabsBaseEvent
	UserAgent string `json:"UserAgent"`
	From      string `json:"From"`
	To        string `json:"To"`
	Length    int    `json:"Length"`
}

// SocketLabsFailedEvent struct
type SocketLabsFailedEvent struct {
	SocketLabsBaseEvent
	BounceStatus   string         `json:"BounceStatus"`
	DiagnosticCode string         `json:"DiagnosticCode"`
	FromAddress    string         `json:"FromAddress"`
	FailureCode    int            `json:"FailureCode"`
	FailureType    string         `json:"FailureType"`
	Reason         string         `json:"Reason"`
	RemoteMta      string         `json:"RemoteMta"`
	Data           SocketLabsData `json:"Data"`
}

// SocketLabsTrackingEvent struct
type SocketLabsTrackingEvent struct {
	SocketLabsBaseEvent
	TrackingType int            `json:"TrackingType"`
	ClientIp     string         `json:"ClientIp"`
	Url          string         `json:"Url"`
	UserAgent    string         `json:"UserAgent"`
	Data         SocketLabsData `json:"Data"`
}

// SocketLabsDeliveredEvent struct
type SocketLabsDeliveredEvent struct {
	SocketLabsBaseEvent
	Response  string         `json:"Response"`
	LocalIp   string         `json:"LocalIp"`
	RemoteMta string         `json:"RemoteMta"`
	Data      SocketLabsData `json:"Data"`
}

// SocketLabsDeferredEvent struct
type SocketLabsDeferredEvent struct {
	SocketLabsBaseEvent
	FromAddress  string         `json:"FromAddress"`
	DeferralCode int            `json:"DeferralCode"`
	Reason       string         `json:"Reason"`
	Data         SocketLabsData `json:"Data"`
}

// SocketLabsQueuedEvent struct
type SocketLabsQueuedEvent struct {
	SocketLabsBaseEvent
	FromAddress string         `json:"FromAddress"`
	Subject     string         `json:"Subject"`
	MessageSize int            `json:"MessageSize"`
	ClientIp    string         `json:"ClientIp"`
	Source      string         `json:"Source"`
	Data        SocketLabsData `json:"Data"`
}

// Supporting struct for Data field
type SocketLabsData struct {
	Meta []SocketLabsMeta `json:"Meta"`
	Tags []string         `json:"Tags"`
}

// Supporting struct for Meta field
type SocketLabsMeta struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}
