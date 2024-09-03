package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"relay-go-consumer/database"

	"github.com/IBM/sarama"
)

// SocketlabsWebhookPayload represents the entire webhook payload
type SocketlabsWebhookPayload struct {
	Headers SocketlabsWebhookHeaders `json:"headers"`
	Body    json.RawMessage          `json:"body"`
}

// WebhookHeaders represents the headers in the webhook payload
type SocketlabsWebhookHeaders struct {
	AcceptEncoding       []string `json:"Accept-Encoding"`
	ContentLength        []string `json:"Content-Length"`
	ContentType          []string `json:"Content-Type"`
	UserAgent            []string `json:"User-Agent"`
	XForwardedFor        []string `json:"X-Forwarded-For"`
	XForwardedHost       []string `json:"X-Forwarded-Host"`
	XForwardedProto      []string `json:"X-Forwarded-Proto"`
	XSocketlabsSignature []string `json:"X-Socketlabs-Signature"`
}

// Tracking type map
var trackingTypeMap = map[int]string{
	0: "Click",
	1: "Open",
	2: "Unsubscribe",
	// Add more mappings as needed
}

// SocketLabsBaseEvent struct for common fields
type SocketLabsBaseEvent struct {
	Type         string    `json:"Type"`
	DateTime     time.Time `json:"DateTime"`
	MailingId    string    `json:"MailingId"`
	MessageId    string    `json:"MessageId"`
	Address      string    `json:"Address"`
	ServerId     int       `json:"ServerId"`
	SubaccountId int       `json:"SubaccountId"`
	IpPoolId     int       `json:"IpPoolId"`
	SecretKey    string    `json:"SecretKey"`
	TrackingType int       `json:"TrackingType"`
	DeferralCode int       `json:"DeferralCode"`
	Reason       string    `json:"Reason"`
	FailureType  string    `json:"FailureType"`
}

func ProcessSocketLabsEvents(msg *sarama.ConsumerMessage) {
	var payload SocketlabsWebhookPayload
	err := json.Unmarshal(msg.Value, &payload)
	if err != nil {
		log.Printf("Failed to unmarshal message: %v\n", err)
		return
	}

	var baseEvent SocketLabsBaseEvent
	err = json.Unmarshal(payload.Body, &baseEvent)
	if err != nil {
		log.Printf("Failed to unmarshal base event: %v\n", err)
		return
	}

	// See Decoding Function to reverse this and ID the sender based on Secret Key
	if baseEvent.MessageId == "" {
		baseEvent.MessageId = generateMessageID(baseEvent.SecretKey, baseEvent.ServerId)
	}

	standardizedEvent := standardizeSocketLabsEvent(baseEvent, payload.Headers)
	err = saveStandardizedSocketLabsEvent(standardizedEvent)
	if err != nil {
		log.Printf("Failed to save standardized event: %v\n", err)
	}
}

func standardizeSocketLabsEvent(event SocketLabsBaseEvent, headers SocketlabsWebhookHeaders) StandardizedEvent {
	standardEvent := StandardizedEvent{
		MessageID:     event.MessageId,
		Provider:      "socketlabs",
		Processed:     true,
		ProcessedTime: event.DateTime.Unix(),
	}

	switch event.Type {
	case "Delivered":
		standardEvent.Delivered = true
		deliveredTime := event.DateTime.Unix()
		standardEvent.DeliveredTime = &deliveredTime
	case "Failed":
		standardEvent.Bounce = true
		bounceTime := event.DateTime.Unix()
		standardEvent.BounceTime = &bounceTime
		if event.FailureType == "Suppressed" {
			standardEvent.Dropped = true
			standardEvent.DroppedTime = &bounceTime
			standardEvent.DroppedReason = event.FailureType
		}
		// You might want to add more logic here to determine the bounce type
	case "Complaint":
		standardEvent.Dropped = true
		droppedTime := event.DateTime.Unix()
		standardEvent.DroppedTime = &droppedTime
		standardEvent.DroppedReason = "Complaint"
	case "Deferred":
		standardEvent.Deferred = true
		standardEvent.DeferredCount = 1
		deferralTime := event.DateTime.Unix()
		standardEvent.LastDeferralTime = &deferralTime
		// You might want to store the DeferralCode and Reason in a structured format
		// For example, you could use a JSON string:
		deferralInfo := fmt.Sprintf(`{"code":%d,"reason":"%s"}`, event.DeferralCode, event.Reason)
		standardEvent.DroppedReason = deferralInfo
	}

	// Handle tracking types
	if trackingType, ok := trackingTypeMap[event.TrackingType]; ok {
		switch trackingType {
		case "Open":
			standardEvent.Open = true
			standardEvent.OpenCount = 1
			openTime := event.DateTime.Unix()
			standardEvent.LastOpenTime = &openTime
			// Assuming all opens from SocketLabs are unique
			standardEvent.UniqueOpen = true
			standardEvent.UniqueOpenTime = &openTime
		case "Click":
			// You might want to add click-specific logic here if needed
		case "Unsubscribe":
			// You might want to add unsubscribe-specific logic here if needed
		}
	}

	return standardEvent
}

func saveStandardizedSocketLabsEvent(event StandardizedEvent) error {
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
            dropped = $10,
            dropped_time = COALESCE($11, dropped_time),
            dropped_reason = COALESCE($12, dropped_reason),
            open = $13,
            open_count = open_count + $14,
            last_open_time = COALESCE($15, last_open_time),
            unique_open = $16,
            unique_open_time = COALESCE($17, unique_open_time),
            deferred = $18,
            deferred_count = deferred_count + $19,
            last_deferral_time = COALESCE($20, last_deferral_time)
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
		event.Dropped,
		event.DroppedTime,
		event.DroppedReason,
		event.Open,
		event.OpenCount,
		event.LastOpenTime,
		event.UniqueOpen,
		event.UniqueOpenTime,
		event.Deferred,
		event.DeferredCount,
		event.LastDeferralTime,
	).Scan(&updatedMessageID)

	if err == sql.ErrNoRows {
		// If no existing record was updated, insert a new one
		insertStmt, err := db.Prepare(`
            INSERT INTO events (
                message_id, provider, processed, processed_time, delivered, delivered_time,
                bounce, bounce_type, bounce_time, dropped, dropped_time, dropped_reason,
                open, open_count, last_open_time, unique_open, unique_open_time,
                deferred, deferred_count, last_deferral_time
            ) VALUES (
                $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
                $18, $19, $20
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
			event.Dropped,
			event.DroppedTime,
			event.DroppedReason,
			event.Open,
			event.OpenCount,
			event.LastOpenTime,
			event.UniqueOpen,
			event.UniqueOpenTime,
			event.Deferred,
			event.DeferredCount,
			event.LastDeferralTime,
		)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func generateMessageID(secretKey string, serverID int) string {
	// Use the current timestamp to ensure uniqueness
	timestamp := time.Now().UnixNano()

	// Create a unique identifier based on SecretKey, ServerID, and timestamp
	uniqueString := fmt.Sprintf("%s:%d:%d", secretKey, serverID, timestamp)

	// Use base64 encoding to create a URL-safe string
	encoded := base64.URLEncoding.EncodeToString([]byte(uniqueString))

	// Trim any padding characters
	return strings.TrimRight(encoded, "=")
}

func decodeMessageID(messageID string) (string, int, int64, error) {
	// Add back the padding if necessary
	if len(messageID)%4 != 0 {
		messageID += strings.Repeat("=", 4-len(messageID)%4)
	}

	// Decode the base64 string
	decoded, err := base64.URLEncoding.DecodeString(messageID)
	if err != nil {
		return "", 0, 0, err
	}

	// Split the decoded string into its components
	parts := strings.Split(string(decoded), ":")
	if len(parts) != 3 {
		return "", 0, 0, fmt.Errorf("invalid message ID format")
	}

	secretKey := parts[0]
	serverID, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid server ID in message ID")
	}

	timestamp, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid timestamp in message ID")
	}

	return secretKey, serverID, timestamp, nil
}

// package main

// import (
// 	"encoding/json"
// 	"log"
// 	"strings"
// 	"time"

// 	"relay-go-consumer/database"

// 	"github.com/IBM/sarama"
// 	"github.com/lib/pq"
// )

// // SocketlabsWebhookPayload represents the entire webhook payload
// type SocketlabsWebhookPayload struct {
// 	Headers SocketlabsWebhookHeaders `json:"headers"`
// 	Body    json.RawMessage          `json:"body"`
// }

// // WebhookHeaders represents the headers in the webhook payload
// type SocketlabsWebhookHeaders struct {
// 	AcceptEncoding       []string `json:"Accept-Encoding"`
// 	ContentLength        []string `json:"Content-Length"`
// 	ContentType          []string `json:"Content-Type"`
// 	UserAgent            []string `json:"User-Agent"`
// 	XForwardedFor        []string `json:"X-Forwarded-For"`
// 	XForwardedHost       []string `json:"X-Forwarded-Host"`
// 	XForwardedProto      []string `json:"X-Forwarded-Proto"`
// 	XSocketlabsSignature []string `json:"X-Socketlabs-Signature"`
// }

// // SocketLabsEventUnmarshaler interface
// type SocketLabsEventUnmarshaler interface {
// 	UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error
// }

// // Tracking type map
// var trackingTypeMap = map[int]string{
// 	0: "Click",
// 	1: "Open",
// 	2: "Unsubscribe",
// 	// Add more mappings as needed
// }

// // SocketLabsBaseEvent struct for common fields
// type SocketLabsBaseEvent struct {
// 	Type         string    `json:"Type"`
// 	DateTime     time.Time `json:"DateTime"`
// 	MailingId    string    `json:"MailingId"`
// 	MessageId    string    `json:"MessageId"`
// 	Address      string    `json:"Address"`
// 	ServerId     int       `json:"ServerId"`
// 	SubaccountId int       `json:"SubaccountId"`
// 	IpPoolId     int       `json:"IpPoolId"`
// 	SecretKey    string    `json:"SecretKey"`
// }

// func (e *SocketLabsBaseEvent) UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error {
// 	if err := json.Unmarshal(data, e); err != nil {
// 		return err
// 	}
// 	return e.saveToDatabase(data, headers)
// }

// func (e *SocketLabsBaseEvent) saveToDatabase(eventData []byte, headers SocketlabsWebhookHeaders) error {
// 	database.InitDB()
// 	db := database.GetDB()

// 	stmt, err := db.Prepare(`
// 		INSERT INTO socketlabs_events (
// 			event_type, date_time, mailing_id, message_id, address, server_id, subaccount_id,
// 			ip_pool_id, secret_key, event_data,
// 			accept_encoding, content_length, content_type, user_agent, x_forwarded_for,
// 			x_forwarded_host, x_forwarded_proto, x_socketlabs_signature, timestamp
// 		) VALUES (
// 			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
// 		)
// 	`)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()

// 	currentTimestamp := time.Now().Unix()

// 	_, err = stmt.Exec(
// 		strings.ToLower(e.Type),
// 		e.DateTime,
// 		e.MailingId,
// 		e.MessageId,
// 		e.Address,
// 		e.ServerId,
// 		e.SubaccountId,
// 		e.IpPoolId,
// 		e.SecretKey,
// 		string(eventData),
// 		pq.Array(headers.AcceptEncoding),
// 		pq.Array(headers.ContentLength),
// 		pq.Array(headers.ContentType),
// 		pq.Array(headers.UserAgent),
// 		pq.Array(headers.XForwardedFor),
// 		pq.Array(headers.XForwardedHost),
// 		pq.Array(headers.XForwardedProto),
// 		pq.Array(headers.XSocketlabsSignature),
// 		currentTimestamp,
// 	)

// 	return err
// }

// // SocketLabsTrackingEvent struct for Click events
// type SocketLabsTrackingEvent struct {
// 	SocketLabsBaseEvent
// 	TrackingType int            `json:"TrackingType"`
// 	ClientIp     string         `json:"ClientIp"`
// 	Url          string         `json:"Url"`
// 	UserAgent    string         `json:"UserAgent"`
// 	Data         SocketLabsData `json:"Data"`
// }

// func (e *SocketLabsTrackingEvent) UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error {
// 	if err := json.Unmarshal(data, e); err != nil {
// 		return err
// 	}

// 	// Handle TrackingType mapping for SocketLabsTrackingEvent
// 	trackingTypeString := e.GetTrackingTypeString()

// 	// Update the Type field with the specific tracking type
// 	e.Type = trackingTypeString

// 	return e.SocketLabsBaseEvent.saveToDatabase(data, headers)
// }

// // SocketLabsComplaintEvent struct for Complaint events
// type SocketLabsComplaintEvent struct {
// 	SocketLabsBaseEvent
// 	UserAgent string `json:"UserAgent"`
// 	From      string `json:"From"`
// 	To        string `json:"To"`
// 	Length    int    `json:"Length"`
// }

// func (e *SocketLabsComplaintEvent) UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error {
// 	if err := json.Unmarshal(data, e); err != nil {
// 		return err
// 	}
// 	return e.SocketLabsBaseEvent.saveToDatabase(data, headers)
// }

// // SocketLabsFailedEvent struct for Failed events
// type SocketLabsFailedEvent struct {
// 	SocketLabsBaseEvent
// 	BounceStatus   string         `json:"BounceStatus"`
// 	DiagnosticCode string         `json:"DiagnosticCode"`
// 	FromAddress    string         `json:"FromAddress"`
// 	FailureCode    int            `json:"FailureCode"`
// 	FailureType    string         `json:"FailureType"`
// 	Reason         string         `json:"Reason"`
// 	RemoteMta      string         `json:"RemoteMta"`
// 	Data           SocketLabsData `json:"Data"`
// }

// func (e *SocketLabsFailedEvent) UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error {
// 	if err := json.Unmarshal(data, e); err != nil {
// 		return err
// 	}
// 	return e.SocketLabsBaseEvent.saveToDatabase(data, headers)
// }

// // SocketLabsDeliveredEvent struct for Delivered events
// type SocketLabsDeliveredEvent struct {
// 	SocketLabsBaseEvent
// 	Response  string         `json:"Response"`
// 	LocalIp   string         `json:"LocalIp"`
// 	RemoteMta string         `json:"RemoteMta"`
// 	Data      SocketLabsData `json:"Data"`
// }

// func (e *SocketLabsDeliveredEvent) UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error {
// 	if err := json.Unmarshal(data, e); err != nil {
// 		return err
// 	}
// 	return e.SocketLabsBaseEvent.saveToDatabase(data, headers)
// }

// // SocketLabsQueuedEvent struct for Queued events
// type SocketLabsQueuedEvent struct {
// 	SocketLabsBaseEvent
// 	FromAddress string         `json:"FromAddress"`
// 	Subject     string         `json:"Subject"`
// 	MessageSize int            `json:"MessageSize"`
// 	ClientIp    string         `json:"ClientIp"`
// 	Source      string         `json:"Source"`
// 	Data        SocketLabsData `json:"Data"`
// }

// func (e *SocketLabsQueuedEvent) UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error {
// 	if err := json.Unmarshal(data, e); err != nil {
// 		return err
// 	}
// 	return e.SocketLabsBaseEvent.saveToDatabase(data, headers)
// }

// // SocketLabsDeferredEvent struct for Deferred events
// type SocketLabsDeferredEvent struct {
// 	SocketLabsBaseEvent
// 	FromAddress  string         `json:"FromAddress"`
// 	DeferralCode int            `json:"DeferralCode"`
// 	Reason       string         `json:"Reason"`
// 	Data         SocketLabsData `json:"Data"`
// }

// func (e *SocketLabsDeferredEvent) UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error {
// 	if err := json.Unmarshal(data, e); err != nil {
// 		return err
// 	}
// 	return e.SocketLabsBaseEvent.saveToDatabase(data, headers)
// }

// // SocketLabsData struct for the Data field
// type SocketLabsData struct {
// 	Meta []SocketLabsMeta `json:"Meta"`
// 	Tags []string         `json:"Tags"`
// }

// // SocketLabsMeta struct for the Meta field
// type SocketLabsMeta struct {
// 	Key   string `json:"Key"`
// 	Value string `json:"Value"`
// }

// func (e *SocketLabsTrackingEvent) GetTrackingTypeString() string {
// 	if typeName, ok := trackingTypeMap[e.TrackingType]; ok {
// 		return typeName
// 	}
// 	return "Unknown"
// }

// func ProcessSocketLabsEvents(msg *sarama.ConsumerMessage) {
// 	var payload SocketlabsWebhookPayload
// 	err := json.Unmarshal(msg.Value, &payload)
// 	if err != nil {
// 		log.Printf("Failed to unmarshal message: %v", err)
// 		return
// 	}

// 	var baseEvent SocketLabsBaseEvent
// 	err = json.Unmarshal(payload.Body, &baseEvent)
// 	if err != nil {
// 		log.Printf("Failed to unmarshal base event: %v", err)
// 		// log.Printf("Raw payload: %s", string(payload.Body))
// 		return
// 	}

// 	var event SocketLabsEventUnmarshaler

// 	switch baseEvent.Type {
// 	case "Tracking":
// 		event = &SocketLabsTrackingEvent{SocketLabsBaseEvent: baseEvent}
// 	case "Complaint":
// 		event = &SocketLabsComplaintEvent{SocketLabsBaseEvent: baseEvent}
// 	case "Failed":
// 		event = &SocketLabsFailedEvent{SocketLabsBaseEvent: baseEvent}
// 	case "Delivered":
// 		event = &SocketLabsDeliveredEvent{SocketLabsBaseEvent: baseEvent}
// 	case "Queued":
// 		event = &SocketLabsQueuedEvent{SocketLabsBaseEvent: baseEvent}
// 	case "Deferred":
// 		event = &SocketLabsDeferredEvent{SocketLabsBaseEvent: baseEvent}
// 	default:
// 		log.Printf("Unknown event type: %s", baseEvent.Type)
// 		return
// 	}

// 	err = event.UnmarshalSocketLabsEvent(payload.Body, payload.Headers)
// 	if err != nil {
// 		log.Printf("Failed to unmarshal event: %v", err)
// 		// log.Printf("Raw payload: %s", string(payload.Body))
// 		return
// 	}

// 	// fmt.Printf("Event: %+v\n", event)
// }
