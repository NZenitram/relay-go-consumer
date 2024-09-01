package main

import (
	"encoding/json"
	"fmt"

	"relay-go-consumer/database"

	"github.com/IBM/sarama"
	"github.com/lib/pq"
)

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

	var headers WebhookHeaders
	err = json.Unmarshal(payload.Headers, &headers)
	if err != nil {
		fmt.Printf("Failed to unmarshal headers: %v\n", err)
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

		err = event.UnmarshalSendgridEvent(eventData, headers)
		if err != nil {
			fmt.Printf("Failed to unmarshal event: %v\n", err)
			continue
		}

		fmt.Printf("Event: %+v\n", event)
	}
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
	UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error
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

func (e *SendgridEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
	if err := json.Unmarshal(data, e); err != nil {
		return err
	}
	e.Provider = "Sendgrid"

	database.InitDB()
	db := database.GetDB()

	// Prepare SQL statement
	stmt, err := db.Prepare(`
        INSERT INTO SendgridEventWithHeaders (
            provider, email, timestamp, smtp_id, event, category, sg_event_id, sg_message_id,
            accept_encoding, content_length, content_type, user_agent, x_forwarded_for,
            x_forwarded_host, x_forwarded_proto, x_twilio_email_event_webhook_signature,
            x_twilio_email_event_webhook_timestamp
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
        )
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute SQL statement
	_, err = stmt.Exec(
		e.Provider,
		e.Email,
		e.Timestamp,
		e.SMTPID,
		e.Event,
		pq.Array(e.Category),
		e.SGEventID,
		e.SGMessageID,
		pq.Array(headers.AcceptEncoding),
		pq.Array(headers.ContentLength),
		pq.Array(headers.ContentType),
		pq.Array(headers.UserAgent),
		pq.Array(headers.XForwardedFor),
		pq.Array(headers.XForwardedHost),
		pq.Array(headers.XForwardedProto),
		pq.Array(headers.XTwilioEmailEventWebhookSignature),
		pq.Array(headers.XTwilioEmailEventWebhookTimestamp),
	)
	if err != nil {
		return err
	}
	return nil
}

// Specific structs for events with additional fields
type SendgridDeferredEvent struct {
	SendgridEvent
	Response string `json:"response"`
	Attempt  string `json:"attempt"`
}

func (s *SendgridDeferredEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
}

type SendgridDeliveredEvent struct {
	SendgridEvent
	Response string `json:"response"`
}

func (s *SendgridDeliveredEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
}

// SendgridOpenEvent
type SendgridOpenEvent struct {
	SendgridEvent
	UserAgent string `json:"useragent"`
	IP        string `json:"ip"`
}

func (s *SendgridOpenEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
}

// SendgridClickEvent
type SendgridClickEvent struct {
	SendgridOpenEvent
	URL string `json:"url"`
}

func (s *SendgridClickEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
}

// SendgridBounceEvent
type SendgridBounceEvent struct {
	SendgridEvent
	Reason string `json:"reason"`
	Status string `json:"status"`
}

func (s *SendgridBounceEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
}

// SendgridDroppedEvent
type SendgridDroppedEvent struct {
	SendgridEvent
	Reason string `json:"reason"`
	Status string `json:"status"`
}

func (s *SendgridDroppedEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
}

// SendgridGroupUnsubscribeEvent
type SendgridGroupUnsubscribeEvent struct {
	SendgridOpenEvent
	URL        string `json:"url"`
	ASMGroupID int    `json:"asm_group_id"`
}

func (s *SendgridGroupUnsubscribeEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
}

// SendgridGroupResubscribeEvent
type SendgridGroupResubscribeEvent struct {
	SendgridOpenEvent
	URL        string `json:"url"`
	ASMGroupID int    `json:"asm_group_id"`
}

func (s *SendgridGroupResubscribeEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
}

// SendgridProcessedEvent
type SendgridProcessedEvent SendgridEvent

func (s *SendgridProcessedEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return (*SendgridEvent)(s).UnmarshalSendgridEvent(data, headers)
}

// SendgridSpamReportEvent
type SendgridSpamReportEvent SendgridEvent

func (s *SendgridSpamReportEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return (*SendgridEvent)(s).UnmarshalSendgridEvent(data, headers)
}

// SendgridUnsubscribeEvent
type SendgridUnsubscribeEvent SendgridEvent

func (s *SendgridUnsubscribeEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	s.Provider = "Sendgrid"
	return (*SendgridEvent)(s).UnmarshalSendgridEvent(data, headers)
}
