package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/mitchellh/mapstructure"
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
		response, err := client.Send(message)
		if err != nil {
			log.Printf("Failed to send email to %s: %v", to.Email, err)
		} else {
			log.Printf("Email sent to %s. Status Code: %d", to.Email, response.StatusCode)
			fmt.Println(response.Body)
			fmt.Println(response.Headers)
		}
	}
}

// WebhookPayload represents the incoming webhook payload
type WebhookPayload struct {
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

// WebhookHeaders represents the headers extracted from the webhook payload
type WebhookHeaders struct {
	AcceptEncoding                    []string `json:"Accept-Encoding"`
	ContentLength                     []string `json:"Content-Length"`
	ContentType                       []string `json:"Content-Type"`
	UserAgent                         []string `json:"User-Agent"`
	XForwardedFor                     []string `json:"X-Forwarded-For"`
	XForwardedHost                    []string `json:"X-Forwarded-Host"`
	XForwardedProto                   []string `json:"X-Forwarded-Proto"`
	XTwilioEmailEventWebhookSignature []string `json:"X-Twilio-Email-Event-Webhook-Signature"`
	XTwilioEmailEventWebhookTimestamp []string `json:"X-Twilio-Email-Event-Webhook-Timestamp"`
}

// ProcessSendgridEvents processes the Sendgrid events from the webhook payload
func ProcessSendgridEvents(msg *sarama.ConsumerMessage) {
	// Unmarshal the message value into a WebhookPayload struct
	var payload WebhookPayload
	err := json.Unmarshal(msg.Value, &payload)
	if err != nil {
		fmt.Printf("Failed to unmarshal message: %v\n", err)
		return
	}

	// Unmarshal the body into an array of SendgridEvent
	var events []SendgridEvent
	json.Unmarshal([]byte(payload.Body), &events)

	// Process each event based on its type
	for _, event := range events {
		switch event.Event {
		case "processed":
			processedEvent := SendgridProcessedEvent(event)
			// Process the processed event
			fmt.Printf("Processed event: %+v\n", processedEvent)

		case "deferred":
			var deferredEvent SendgridDeferredEvent
			mapstructure.Decode(event, &deferredEvent)
			// Process the deferred event
			fmt.Printf("Deferred event: %+v\n", deferredEvent)

		case "delivered":
			var deliveredEvent SendgridDeliveredEvent
			mapstructure.Decode(event, &deliveredEvent)
			// Process the delivered event
			fmt.Printf("Delivered event: %+v\n", deliveredEvent)

		case "open":
			var openEvent SendgridOpenEvent
			mapstructure.Decode(event, &openEvent)
			// Process the open event
			fmt.Printf("Open event: %+v\n", openEvent)

		case "click":
			var clickEvent SendgridClickEvent
			mapstructure.Decode(event, &clickEvent)
			// Process the click event
			fmt.Printf("Click event: %+v\n", clickEvent)

		case "bounce":
			var bounceEvent SendgridBounceEvent
			mapstructure.Decode(event, &bounceEvent)
			// Process the bounce event
			fmt.Printf("Bounce event: %+v\n", bounceEvent)

		case "dropped":
			var droppedEvent SendgridDroppedEvent
			mapstructure.Decode(event, &droppedEvent)
			// Process the dropped event
			fmt.Printf("Dropped event: %+v\n", droppedEvent)

		case "spamreport":
			spamReportEvent := SendgridSpamReportEvent(event)
			// Process the spam report event
			fmt.Printf("Spam report event: %+v\n", spamReportEvent)

		case "unsubscribe":
			unsubscribeEvent := SendgridUnsubscribeEvent(event)
			// Process the unsubscribe event
			fmt.Printf("Unsubscribe event: %+v\n", unsubscribeEvent)

		case "group_unsubscribe":
			var groupUnsubscribeEvent SendgridGroupUnsubscribeEvent
			mapstructure.Decode(event, &groupUnsubscribeEvent)
			// Process the group unsubscribe event
			fmt.Printf("Group unsubscribe event: %+v\n", groupUnsubscribeEvent)

		case "group_resubscribe":
			var groupResubscribeEvent SendgridGroupResubscribeEvent
			mapstructure.Decode(event, &groupResubscribeEvent)
			// Process the group resubscribe event
			fmt.Printf("Group resubscribe event: %+v\n", groupResubscribeEvent)

		default:
			fmt.Printf("Unknown event type: %s\n", event.Event)
		}
	}
}

func (e *SendgridEvent) UnmarshalJSON(data []byte) error {
	type Alias SendgridEvent
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	e.Provider = "Sendgrid"
	return nil
}

// Base struct for common fields
type SendgridEvent struct {
	Provider    string   `json:"provider"`
	Email       string   `json:"email"`
	Timestamp   int64    `json:"timestamp"`
	SMTPID      string   `json:"smtp-id"`
	Event       string   `json:"event"`
	Category    []string `json:"category"`
	SGEventID   string   `json:"sg_event_id"`
	SGMessageID string   `json:"sg_message_id"`
}

// Specific structs for events with additional fields
type SendgridDeferredEvent struct {
	SendgridEvent
	Response string `json:"response"`
	Attempt  string `json:"attempt"`
}

type SendgridDeliveredEvent struct {
	SendgridEvent
	Response string `json:"response"`
}

type SendgridOpenEvent struct {
	SendgridEvent
	UserAgent string `json:"useragent"`
	IP        string `json:"ip"`
}

type SendgridClickEvent struct {
	SendgridOpenEvent
	URL string `json:"url"`
}

type SendgridBounceEvent struct {
	SendgridEvent
	Reason string `json:"reason"`
	Status string `json:"status"`
}

type SendgridDroppedEvent struct {
	SendgridEvent
	Reason string `json:"reason"`
	Status string `json:"status"`
}

type SendgridGroupUnsubscribeEvent struct {
	SendgridOpenEvent
	URL        string `json:"url"`
	ASMGroupID int    `json:"asm_group_id"`
}

type SendgridGroupResubscribeEvent struct {
	SendgridOpenEvent
	URL        string `json:"url"`
	ASMGroupID int    `json:"asm_group_id"`
}

// Events with no additional fields can use the base Event struct
type SendgridProcessedEvent SendgridEvent
type SendgridSpamReportEvent SendgridEvent
type SendgridUnsubscribeEvent SendgridEvent
