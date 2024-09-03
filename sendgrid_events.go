package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"relay-go-consumer/database"
	"strconv"

	"github.com/IBM/sarama"
)

type EventPayload struct {
	Headers map[string][]string `json:"headers"`
	Body    []json.RawMessage   `json:"body"`
}

type EventBody struct {
	Email         string   `json:"email"`
	Event         string   `json:"event"`
	SGMessageID   string   `json:"sg_message_id"`
	SGMachineOpen bool     `json:"sg_machine_open"`
	Timestamp     int64    `json:"timestamp"`
	Category      []string `json:"category"`
	SGEventID     string   `json:"sg_event_id"`
	SMTPID        string   `json:"smtp-id"`
	BounceType    string   `json:"bounce_type,omitempty"`
	Reason        string   `json:"reason,omitempty"`
}

type StandardizedEvent struct {
	MessageID        string
	Provider         string
	Processed        bool
	ProcessedTime    int64
	Delivered        bool
	DeliveredTime    *int64
	Bounce           bool
	BounceType       string
	BounceTime       *int64
	Deferred         bool
	DeferredCount    int
	LastDeferralTime *int64
	UniqueOpen       bool
	UniqueOpenTime   *int64
	Open             bool
	OpenCount        int
	LastOpenTime     *int64
	Dropped          bool
	DroppedTime      *int64
	DroppedReason    string
}

func ProcessSendgridEvents(msg *sarama.ConsumerMessage) {
	var payload EventPayload
	err := json.Unmarshal(msg.Value, &payload)
	if err != nil {
		fmt.Printf("failed to unmarshal message: %v", err)
	}

	for _, eventData := range payload.Body {
		var eventBody EventBody
		err := json.Unmarshal(eventData, &eventBody)
		if err != nil {
			fmt.Printf("Failed to unmarshal event body: %v\n", err)
			continue
		}

		standardizedEvent := standardizeEvent(eventBody, payload.Headers)
		err = saveStandardizedEvent(standardizedEvent)
		if err != nil {
			fmt.Printf("Failed to save standardized event: %v\n", err)
		}
	}
}

func standardizeEvent(eventBody EventBody, headers map[string][]string) StandardizedEvent {
	processedTime, _ := strconv.ParseInt(headers["X-Twilio-Email-Event-Webhook-Timestamp"][0], 10, 64)

	event := StandardizedEvent{
		MessageID:     eventBody.SGMessageID,
		Provider:      "sendgrid",
		Processed:     true,
		ProcessedTime: processedTime,
	}

	switch eventBody.Event {
	case "delivered":
		event.Delivered = true
		event.DeliveredTime = &eventBody.Timestamp
	case "bounce":
		event.Bounce = true
		event.BounceTime = &eventBody.Timestamp
		event.BounceType = eventBody.BounceType
	case "deferred":
		event.Deferred = true
		event.DeferredCount = 1
		event.LastDeferralTime = &eventBody.Timestamp
	case "open":
		event.Open = true
		event.OpenCount = 1
		event.LastOpenTime = &eventBody.Timestamp
		if !eventBody.SGMachineOpen {
			event.UniqueOpen = true
			event.UniqueOpenTime = &eventBody.Timestamp
		}
	case "dropped":
		event.Dropped = true
		event.DroppedTime = &eventBody.Timestamp
		event.DroppedReason = eventBody.Reason
	}

	return event
}

func saveStandardizedEvent(event StandardizedEvent) error {
	database.InitDB()
	db := database.GetDB()

	// First, try to update an existing record
	stmt, err := db.Prepare(`
        UPDATE events SET
            provider = $2,
            processed = $3,
            processed_time = $4,
            delivered = $5,
            delivered_time = COALESCE($6, delivered_time),
            bounce = $7,
            bounce_type = COALESCE($8, bounce_type),
            bounce_time = COALESCE($9, bounce_time),
            deferred = $10,
            deferred_count = deferred_count + $11,
            last_deferral_time = COALESCE($12, last_deferral_time),
            unique_open = $13,
            unique_open_time = COALESCE($14, unique_open_time),
            open = $15,
            open_count = open_count + $16,
            last_open_time = COALESCE($17, last_open_time),
            dropped = $18,
            dropped_time = COALESCE($19, dropped_time),
            dropped_reason = COALESCE($20, dropped_reason)
        WHERE message_id = $1
        RETURNING message_id
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var updatedMessageID string
	err = stmt.QueryRow(
		event.MessageID,
		event.Provider,
		event.Processed,
		event.ProcessedTime,
		event.Delivered,
		event.DeliveredTime,
		event.Bounce,
		event.BounceType,
		event.BounceTime,
		event.Deferred,
		event.DeferredCount,
		event.LastDeferralTime,
		event.UniqueOpen,
		event.UniqueOpenTime,
		event.Open,
		event.OpenCount,
		event.LastOpenTime,
		event.Dropped,
		event.DroppedTime,
		event.DroppedReason,
	).Scan(&updatedMessageID)

	if err == sql.ErrNoRows {
		// If no existing record was updated, insert a new one
		insertStmt, err := db.Prepare(`
            INSERT INTO events (
                message_id, provider, processed, processed_time, delivered, delivered_time,
                bounce, bounce_type, bounce_time, deferred, deferred_count,
                last_deferral_time, unique_open, unique_open_time, open, open_count, last_open_time,
                dropped, dropped_time, dropped_reason
            ) VALUES (
                $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
            )
        `)
		if err != nil {
			return err
		}
		defer insertStmt.Close()

		_, err = insertStmt.Exec(
			event.MessageID,
			event.Provider,
			event.Processed,
			event.ProcessedTime,
			event.Delivered,
			event.DeliveredTime,
			event.Bounce,
			event.BounceType,
			event.BounceTime,
			event.Deferred,
			event.DeferredCount,
			event.LastDeferralTime,
			event.UniqueOpen,
			event.UniqueOpenTime,
			event.Open,
			event.OpenCount,
			event.LastOpenTime,
			event.Dropped,
			event.DroppedTime,
			event.DroppedReason,
		)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

// package main

// import (
// 	"encoding/json"
// 	"fmt"

// 	"relay-go-consumer/database"

// 	"github.com/IBM/sarama"
// 	"github.com/lib/pq"
// )

// func ProcessSendgridEvents(msg *sarama.ConsumerMessage) {
// 	var payload struct {
// 		Headers json.RawMessage   `json:"headers"`
// 		Body    []json.RawMessage `json:"body"`
// 	}
// 	err := json.Unmarshal(msg.Value, &payload)
// 	if err != nil {
// 		fmt.Printf("Failed to unmarshal message: %v\n", err)
// 		return
// 	}

// 	var headers WebhookHeaders
// 	err = json.Unmarshal(payload.Headers, &headers)
// 	if err != nil {
// 		fmt.Printf("Failed to unmarshal headers: %v\n", err)
// 		return
// 	}

// 	for _, eventData := range payload.Body {
// 		var baseEvent SendgridEvent
// 		err := json.Unmarshal(eventData, &baseEvent)
// 		if err != nil {
// 			fmt.Printf("Failed to unmarshal base event: %v\n", err)
// 			continue
// 		}

// 		var event SendgridEventUnmarshaler

// 		switch baseEvent.Event {
// 		case "processed":
// 			event = &SendgridProcessedEvent{}
// 		case "deferred":
// 			event = &SendgridDeferredEvent{}
// 		case "delivered":
// 			event = &SendgridDeliveredEvent{}
// 		case "open":
// 			event = &SendgridOpenEvent{}
// 		case "click":
// 			event = &SendgridClickEvent{}
// 		case "bounce":
// 			event = &SendgridBounceEvent{}
// 		case "dropped":
// 			event = &SendgridDroppedEvent{}
// 		case "spamreport":
// 			event = &SendgridSpamReportEvent{}
// 		case "unsubscribe":
// 			event = &SendgridUnsubscribeEvent{}
// 		case "group_unsubscribe":
// 			event = &SendgridGroupUnsubscribeEvent{}
// 		case "group_resubscribe":
// 			event = &SendgridGroupResubscribeEvent{}
// 		default:
// 			fmt.Printf("Unknown event type: %s\n", baseEvent.Event)
// 			continue
// 		}

// 		err = event.UnmarshalSendgridEvent(eventData, headers)
// 		if err != nil {
// 			fmt.Printf("Failed to unmarshal event: %v\n", err)
// 			continue
// 		}

// 		fmt.Printf("Event: %+v\n", event)
// 	}
// }

// // WebhookPayload represents the incoming webhook payload
// type WebhookPayload struct {
// 	Headers map[string][]string `json:"headers"`
// 	Body    []SendgridEvent     `json:"body"`
// }

// // WebhookHeaders represents the headers extracted from the webhook payload
// type WebhookHeaders struct {
// 	AcceptEncoding                    []string `json:"Accept-Encoding"`
// 	ContentLength                     []string `json:"Content-Length"`
// 	ContentType                       []string `json:"Content-Type"`
// 	UserAgent                         []string `json:"User-Agent"`
// 	XForwardedFor                     []string `json:"X-Forwarded-For"`
// 	XForwardedHost                    []string `json:"X-Forwarded-Host"`
// 	XForwardedProto                   []string `json:"X-Forwarded-Proto"`
// 	XTwilioEmailEventWebhookSignature []string `json:"X-Twilio-Email-Event-Webhook-Signature"`
// 	XTwilioEmailEventWebhookTimestamp []string `json:"X-Twilio-Email-Event-Webhook-Timestamp"`
// }

// // SendgridEventUnmarshaler interface
// type SendgridEventUnmarshaler interface {
// 	UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error
// }

// // Base struct for common fields
// type SendgridEvent struct {
// 	Provider    string   `json:"provider"`
// 	Email       string   `json:"email"`
// 	Timestamp   int64    `json:"timestamp"`
// 	SMTPID      string   `json:"smtp-id"`
// 	Event       string   `json:"event"`
// 	Category    []string `json:"category"`
// 	SGEventID   string   `json:"sg_event_id"`
// 	SGMessageID string   `json:"sg_message_id"`
// }

// func (e *SendgridEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
// 	if err := json.Unmarshal(data, e); err != nil {
// 		return err
// 	}
// 	e.Provider = "Sendgrid"

// 	database.InitDB()
// 	db := database.GetDB()

// 	// Prepare SQL statement
// 	stmt, err := db.Prepare(`
//         INSERT INTO sendgrid_events (
//             provider, email, timestamp, smtp_id, event, category, sg_event_id, sg_message_id,
//             accept_encoding, content_length, content_type, user_agent, x_forwarded_for,
//             x_forwarded_host, x_forwarded_proto, x_twilio_email_event_webhook_signature,
//             x_twilio_email_event_webhook_timestamp
//         ) VALUES (
//             $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
//         )
//     `)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()

// 	// Execute SQL statement
// 	_, err = stmt.Exec(
// 		e.Provider,
// 		e.Email,
// 		e.Timestamp,
// 		e.SMTPID,
// 		e.Event,
// 		pq.Array(e.Category),
// 		e.SGEventID,
// 		e.SGMessageID,
// 		pq.Array(headers.AcceptEncoding),
// 		pq.Array(headers.ContentLength),
// 		pq.Array(headers.ContentType),
// 		pq.Array(headers.UserAgent),
// 		pq.Array(headers.XForwardedFor),
// 		pq.Array(headers.XForwardedHost),
// 		pq.Array(headers.XForwardedProto),
// 		pq.Array(headers.XTwilioEmailEventWebhookSignature),
// 		pq.Array(headers.XTwilioEmailEventWebhookTimestamp),
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // Specific structs for events with additional fields
// type SendgridDeferredEvent struct {
// 	SendgridEvent
// 	Response string `json:"response"`
// 	Attempt  string `json:"attempt"`
// }

// func (s *SendgridDeferredEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
// 	if err := json.Unmarshal(data, s); err != nil {
// 		return err
// 	}
// 	s.Provider = "Sendgrid"
// 	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
// }

// type SendgridDeliveredEvent struct {
// 	SendgridEvent
// 	Response string `json:"response"`
// }

// func (s *SendgridDeliveredEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
// 	if err := json.Unmarshal(data, s); err != nil {
// 		return err
// 	}
// 	s.Provider = "Sendgrid"
// 	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
// }

// // SendgridOpenEvent
// type SendgridOpenEvent struct {
// 	SendgridEvent
// 	UserAgent string `json:"useragent"`
// 	IP        string `json:"ip"`
// }

// func (s *SendgridOpenEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
// 	if err := json.Unmarshal(data, s); err != nil {
// 		return err
// 	}
// 	s.Provider = "Sendgrid"
// 	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
// }

// // SendgridClickEvent
// type SendgridClickEvent struct {
// 	SendgridOpenEvent
// 	URL string `json:"url"`
// }

// func (s *SendgridClickEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
// 	if err := json.Unmarshal(data, s); err != nil {
// 		return err
// 	}
// 	s.Provider = "Sendgrid"
// 	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
// }

// // SendgridBounceEvent
// type SendgridBounceEvent struct {
// 	SendgridEvent
// 	Reason string `json:"reason"`
// 	Status string `json:"status"`
// }

// func (s *SendgridBounceEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
// 	if err := json.Unmarshal(data, s); err != nil {
// 		return err
// 	}
// 	s.Provider = "Sendgrid"
// 	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
// }

// // SendgridDroppedEvent
// type SendgridDroppedEvent struct {
// 	SendgridEvent
// 	Reason string `json:"reason"`
// 	Status string `json:"status"`
// }

// func (s *SendgridDroppedEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
// 	if err := json.Unmarshal(data, s); err != nil {
// 		return err
// 	}
// 	s.Provider = "Sendgrid"
// 	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
// }

// // SendgridGroupUnsubscribeEvent
// type SendgridGroupUnsubscribeEvent struct {
// 	SendgridOpenEvent
// 	URL        string `json:"url"`
// 	ASMGroupID int    `json:"asm_group_id"`
// }

// func (s *SendgridGroupUnsubscribeEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
// 	if err := json.Unmarshal(data, s); err != nil {
// 		return err
// 	}
// 	s.Provider = "Sendgrid"
// 	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
// }

// // SendgridGroupResubscribeEvent
// type SendgridGroupResubscribeEvent struct {
// 	SendgridOpenEvent
// 	URL        string `json:"url"`
// 	ASMGroupID int    `json:"asm_group_id"`
// }

// func (s *SendgridGroupResubscribeEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
// 	if err := json.Unmarshal(data, s); err != nil {
// 		return err
// 	}
// 	s.Provider = "Sendgrid"
// 	return s.SendgridEvent.UnmarshalSendgridEvent(data, headers)
// }

// // SendgridProcessedEvent
// type SendgridProcessedEvent SendgridEvent

// func (s *SendgridProcessedEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
// 	if err := json.Unmarshal(data, s); err != nil {
// 		return err
// 	}
// 	s.Provider = "Sendgrid"
// 	return (*SendgridEvent)(s).UnmarshalSendgridEvent(data, headers)
// }

// // SendgridSpamReportEvent
// type SendgridSpamReportEvent SendgridEvent

// func (s *SendgridSpamReportEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
// 	if err := json.Unmarshal(data, s); err != nil {
// 		return err
// 	}
// 	s.Provider = "Sendgrid"
// 	return (*SendgridEvent)(s).UnmarshalSendgridEvent(data, headers)
// }

// // SendgridUnsubscribeEvent
// type SendgridUnsubscribeEvent SendgridEvent

// func (s *SendgridUnsubscribeEvent) UnmarshalSendgridEvent(data []byte, headers WebhookHeaders) error {
// 	if err := json.Unmarshal(data, s); err != nil {
// 		return err
// 	}
// 	s.Provider = "Sendgrid"
// 	return (*SendgridEvent)(s).UnmarshalSendgridEvent(data, headers)
// }
