package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
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

func ProcessSendgridEvents(msg *sarama.ConsumerMessage) {
	var payload struct {
		Headers json.RawMessage   `json:"headers"`
		Body    []json.RawMessage `json:"body"`
	}
	err := json.Unmarshal(msg.Value, &payload)
	if err != nil {
		fmt.Printf("Failed to unmarshal message: %v\n", err)
		return
	}

	for _, eventData := range payload.Body {
		var baseEvent SendgridEvent
		err := json.Unmarshal(eventData, &baseEvent)
		if err != nil {
			fmt.Printf("Failed to unmarshal base event: %v\n", err)
			continue
		}

		var event SendgridEventUnmarshaler

		switch baseEvent.Event {
		case "processed":
			event = &SendgridProcessedEvent{}
		case "deferred":
			event = &SendgridDeferredEvent{}
		case "delivered":
			event = &SendgridDeliveredEvent{}
		case "open":
			event = &SendgridOpenEvent{}
		case "click":
			event = &SendgridClickEvent{}
		case "bounce":
			event = &SendgridBounceEvent{}
		case "dropped":
			event = &SendgridDroppedEvent{}
		case "spamreport":
			event = &SendgridSpamReportEvent{}
		case "unsubscribe":
			event = &SendgridUnsubscribeEvent{}
		case "group_unsubscribe":
			event = &SendgridGroupUnsubscribeEvent{}
		case "group_resubscribe":
			event = &SendgridGroupResubscribeEvent{}
		default:
			fmt.Printf("Unknown event type: %s\n", baseEvent.Event)
			continue
		}

		err = event.UnmarshalSendgridEvent(eventData)
		if err != nil {
			fmt.Printf("Failed to unmarshal event: %v\n", err)
			continue
		}

		fmt.Printf("Event: %+v\n", event)
		saveToDatabase(event)
	}
}

// Placeholder function to save events to a database
func saveToDatabase(event interface{}) {
	// TODO: Implement database saving logic
	fmt.Printf("Saving event to database: %+v\n", event)
}

// WebhookPayload represents the incoming webhook payload
type WebhookPayload struct {
	Headers map[string][]string `json:"headers"`
	Body    []SendgridEvent     `json:"body"`
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

// SendgridEventUnmarshaler interface
type SendgridEventUnmarshaler interface {
	UnmarshalSendgridEvent(data []byte) error
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

func (e *SendgridEvent) UnmarshalSendgridEvent(data []byte) error {
	if err := json.Unmarshal(data, e); err != nil {
		return err
	}
	e.Provider = "Sendgrid"
	return nil
}

// Specific structs for events with additional fields
type SendgridDeferredEvent struct {
	SendgridEvent
	Response string `json:"response"`
	Attempt  string `json:"attempt"`
}

func (s *SendgridDeferredEvent) UnmarshalSendgridEvent(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return nil
}

// Implement UnmarshalSendgridEvent for other event types similarly
// For example:

type SendgridDeliveredEvent struct {
	SendgridEvent
	Response string `json:"response"`
}

func (s *SendgridDeliveredEvent) UnmarshalSendgridEvent(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return nil
}

// SendgridOpenEvent
type SendgridOpenEvent struct {
	SendgridEvent
	UserAgent string `json:"useragent"`
	IP        string `json:"ip"`
}

func (s *SendgridOpenEvent) UnmarshalSendgridEvent(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return nil
}

// SendgridClickEvent
type SendgridClickEvent struct {
	SendgridOpenEvent
	URL string `json:"url"`
}

func (s *SendgridClickEvent) UnmarshalSendgridEvent(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return nil
}

// SendgridBounceEvent
type SendgridBounceEvent struct {
	SendgridEvent
	Reason string `json:"reason"`
	Status string `json:"status"`
}

func (s *SendgridBounceEvent) UnmarshalSendgridEvent(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return nil
}

// SendgridDroppedEvent
type SendgridDroppedEvent struct {
	SendgridEvent
	Reason string `json:"reason"`
	Status string `json:"status"`
}

func (s *SendgridDroppedEvent) UnmarshalSendgridEvent(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return nil
}

// SendgridGroupUnsubscribeEvent
type SendgridGroupUnsubscribeEvent struct {
	SendgridOpenEvent
	URL        string `json:"url"`
	ASMGroupID int    `json:"asm_group_id"`
}

func (s *SendgridGroupUnsubscribeEvent) UnmarshalSendgridEvent(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return nil
}

// SendgridGroupResubscribeEvent
type SendgridGroupResubscribeEvent struct {
	SendgridOpenEvent
	URL        string `json:"url"`
	ASMGroupID int    `json:"asm_group_id"`
}

func (s *SendgridGroupResubscribeEvent) UnmarshalSendgridEvent(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return nil
}

// SendgridProcessedEvent
type SendgridProcessedEvent SendgridEvent

func (s *SendgridProcessedEvent) UnmarshalSendgridEvent(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return nil
}

// SendgridSpamReportEvent
type SendgridSpamReportEvent SendgridEvent

func (s *SendgridSpamReportEvent) UnmarshalSendgridEvent(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return nil
}

// SendgridUnsubscribeEvent
type SendgridUnsubscribeEvent SendgridEvent

func (s *SendgridUnsubscribeEvent) UnmarshalSendgridEvent(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return nil
}
