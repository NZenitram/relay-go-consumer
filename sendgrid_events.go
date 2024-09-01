package main

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
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
