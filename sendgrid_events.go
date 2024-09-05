package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/sendgrid/sendgrid-go/helpers/eventwebhook"
)

type SendgridHeaders struct {
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

type SendgridWebhookPayload struct {
	Headers map[string][]string `json:"headers"`
	Body    json.RawMessage     `json:"body"`
}

type EventPayload struct {
	Headers SendgridHeaders   `json:"headers"`
	Body    []json.RawMessage `json:"body"`
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

func ProcessSendgridEvents(msg *sarama.ConsumerMessage) {
	var message SendgridWebhookPayload
	msgErr := json.Unmarshal(msg.Value, &message)
	if msgErr != nil {
		fmt.Printf("failed to unmarshal message: %v", msgErr)
	}

	var payload EventPayload
	err := json.Unmarshal(msg.Value, &payload)
	if err != nil {
		fmt.Printf("failed to unmarshal message: %v", err)
	}

	if err != nil {
		fmt.Printf("Failed to verify webhook and find user: %v\n", err)
		return
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

func standardizeEvent(eventBody EventBody, headers SendgridHeaders) StandardizedEvent {
	processedTime, _ := strconv.ParseInt(headers.XTwilioEmailEventWebhookTimestamp[0], 10, 64)

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

func verifySendgridWebhookAndFindUser(db *sql.DB, message SendgridWebhookPayload, rawBody []byte) (int, error) {
	// if len(payload.Headers.XTwilioEmailEventWebhookSignature) == 0 {
	// 	return 0, fmt.Errorf("missing X-Twilio-Email-Event-Webhook-Signature header")
	// }
	s := message.Headers["X-Twilio-Email-Event-Webhook-Signature"][0]

	// if len(payload.Headers.XTwilioEmailEventWebhookTimestamp) == 0 {
	// 	return 0, fmt.Errorf("missing X-Twilio-Email-Event-Webhook-Timestamp header")
	// }
	ts := message.Headers["X-Twilio-Email-Event-Webhook-Timestamp"][0]

	rows, err := db.Query("SELECT user_id, sendgrid_verification_key FROM email_service_providers WHERE provider_name = 'sendgrid' AND sendgrid_verification_key IS NOT NULL")
	if err != nil {
		return 0, fmt.Errorf("failed to query database: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		var publicKey string
		err := rows.Scan(&userID, &publicKey)
		if err != nil {
			return 0, fmt.Errorf("failed to scan row: %v", err)
		}

		// Takes a base64 ECDSA public key and converts it into the ECDSA Public Key type

		ecdaKey, err := eventwebhook.ConvertPublicKeyBase64ToECDSA(publicKey)
		if err != nil {
			log.Printf("Cannot convert public key: %v", err)
			continue
		}

		valid, err := eventwebhook.VerifySignature(ecdaKey, rawBody, s, ts)
		if err != nil {
			continue // Skip this key if verification fails
		}

		if valid {
			return userID, nil // We found a matching user
		}
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("error iterating over rows: %v", err)
	}

	return 0, fmt.Errorf("no matching user found for the given webhook signature")
}

func associateSendgridEventWithUser(db *sql.DB, messageID string, userID int) error {
	// Insert the association into the message_user_associations table
	_, err := db.Exec(`
        INSERT INTO message_user_associations (message_id, user_id, esp_id, provider)
        VALUES ($1, $2, (SELECT esp_id FROM email_service_providers WHERE user_id = $2 AND provider_name = 'sendgrid'), 'sendgrid')
        ON CONFLICT (message_id, provider) DO NOTHING
    `, messageID, userID)
	if err != nil {
		return fmt.Errorf("failed to insert message association: %v", err)
	}

	return nil
}

func removeWhitespace(data []byte) []byte {
	return bytes.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, data)
}

func formatRawBody(body json.RawMessage) []byte {
	// Check if the body is already an array
	if len(body) > 0 && body[0] == '[' {
		// It's already an array, just ensure it ends with a newline
		if !bytes.HasSuffix(body, []byte("\r\n")) {
			return append(body, '\r', '\n')
		}
		return body
	}

	// If it's not an array, wrap it in an array
	var formatted bytes.Buffer
	formatted.WriteString("[")
	formatted.Write(body)
	formatted.WriteString("]\r\n")
	return formatted.Bytes()
}
