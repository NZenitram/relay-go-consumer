package main

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"relay-go-consumer/database"

	"github.com/IBM/sarama"
	"github.com/lib/pq"
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

// SocketLabsEventUnmarshaler interface
type SocketLabsEventUnmarshaler interface {
	UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error
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
}

func (e *SocketLabsBaseEvent) UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error {
	if err := json.Unmarshal(data, e); err != nil {
		return err
	}
	return e.saveToDatabase(data, headers)
}

func (e *SocketLabsBaseEvent) saveToDatabase(eventData []byte, headers SocketlabsWebhookHeaders) error {
	database.InitDB()
	db := database.GetDB()

	stmt, err := db.Prepare(`
		INSERT INTO socketlabs_events (
			event_type, date_time, mailing_id, message_id, address, server_id, subaccount_id, 
			ip_pool_id, secret_key, event_data,
			accept_encoding, content_length, content_type, user_agent, x_forwarded_for,
			x_forwarded_host, x_forwarded_proto, x_socketlabs_signature, timestamp
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	currentTimestamp := time.Now().Unix()

	_, err = stmt.Exec(
		strings.ToLower(e.Type),
		e.DateTime,
		e.MailingId,
		e.MessageId,
		e.Address,
		e.ServerId,
		e.SubaccountId,
		e.IpPoolId,
		e.SecretKey,
		string(eventData),
		pq.Array(headers.AcceptEncoding),
		pq.Array(headers.ContentLength),
		pq.Array(headers.ContentType),
		pq.Array(headers.UserAgent),
		pq.Array(headers.XForwardedFor),
		pq.Array(headers.XForwardedHost),
		pq.Array(headers.XForwardedProto),
		pq.Array(headers.XSocketlabsSignature),
		currentTimestamp,
	)

	return err
}

// SocketLabsTrackingEvent struct for Click events
type SocketLabsTrackingEvent struct {
	SocketLabsBaseEvent
	TrackingType int            `json:"TrackingType"`
	ClientIp     string         `json:"ClientIp"`
	Url          string         `json:"Url"`
	UserAgent    string         `json:"UserAgent"`
	Data         SocketLabsData `json:"Data"`
}

func (e *SocketLabsTrackingEvent) UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error {
	if err := json.Unmarshal(data, e); err != nil {
		return err
	}

	// Handle TrackingType mapping for SocketLabsTrackingEvent
	trackingTypeString := e.GetTrackingTypeString()

	// Update the Type field with the specific tracking type
	e.Type = trackingTypeString

	return e.SocketLabsBaseEvent.saveToDatabase(data, headers)
}

// SocketLabsComplaintEvent struct for Complaint events
type SocketLabsComplaintEvent struct {
	SocketLabsBaseEvent
	UserAgent string `json:"UserAgent"`
	From      string `json:"From"`
	To        string `json:"To"`
	Length    int    `json:"Length"`
}

func (e *SocketLabsComplaintEvent) UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error {
	if err := json.Unmarshal(data, e); err != nil {
		return err
	}
	return e.SocketLabsBaseEvent.saveToDatabase(data, headers)
}

// SocketLabsFailedEvent struct for Failed events
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

func (e *SocketLabsFailedEvent) UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error {
	if err := json.Unmarshal(data, e); err != nil {
		return err
	}
	return e.SocketLabsBaseEvent.saveToDatabase(data, headers)
}

// SocketLabsDeliveredEvent struct for Delivered events
type SocketLabsDeliveredEvent struct {
	SocketLabsBaseEvent
	Response  string         `json:"Response"`
	LocalIp   string         `json:"LocalIp"`
	RemoteMta string         `json:"RemoteMta"`
	Data      SocketLabsData `json:"Data"`
}

func (e *SocketLabsDeliveredEvent) UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error {
	if err := json.Unmarshal(data, e); err != nil {
		return err
	}
	return e.SocketLabsBaseEvent.saveToDatabase(data, headers)
}

// SocketLabsQueuedEvent struct for Queued events
type SocketLabsQueuedEvent struct {
	SocketLabsBaseEvent
	FromAddress string         `json:"FromAddress"`
	Subject     string         `json:"Subject"`
	MessageSize int            `json:"MessageSize"`
	ClientIp    string         `json:"ClientIp"`
	Source      string         `json:"Source"`
	Data        SocketLabsData `json:"Data"`
}

func (e *SocketLabsQueuedEvent) UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error {
	if err := json.Unmarshal(data, e); err != nil {
		return err
	}
	return e.SocketLabsBaseEvent.saveToDatabase(data, headers)
}

// SocketLabsDeferredEvent struct for Deferred events
type SocketLabsDeferredEvent struct {
	SocketLabsBaseEvent
	FromAddress  string         `json:"FromAddress"`
	DeferralCode int            `json:"DeferralCode"`
	Reason       string         `json:"Reason"`
	Data         SocketLabsData `json:"Data"`
}

func (e *SocketLabsDeferredEvent) UnmarshalSocketLabsEvent(data []byte, headers SocketlabsWebhookHeaders) error {
	if err := json.Unmarshal(data, e); err != nil {
		return err
	}
	return e.SocketLabsBaseEvent.saveToDatabase(data, headers)
}

// SocketLabsData struct for the Data field
type SocketLabsData struct {
	Meta []SocketLabsMeta `json:"Meta"`
	Tags []string         `json:"Tags"`
}

// SocketLabsMeta struct for the Meta field
type SocketLabsMeta struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

func (e *SocketLabsTrackingEvent) GetTrackingTypeString() string {
	if typeName, ok := trackingTypeMap[e.TrackingType]; ok {
		return typeName
	}
	return "Unknown"
}

func ProcessSocketLabsEvents(msg *sarama.ConsumerMessage) {
	var payload SocketlabsWebhookPayload
	err := json.Unmarshal(msg.Value, &payload)
	if err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	var baseEvent SocketLabsBaseEvent
	err = json.Unmarshal(payload.Body, &baseEvent)
	if err != nil {
		log.Printf("Failed to unmarshal base event: %v", err)
		// log.Printf("Raw payload: %s", string(payload.Body))
		return
	}

	var event SocketLabsEventUnmarshaler

	switch baseEvent.Type {
	case "Tracking":
		event = &SocketLabsTrackingEvent{SocketLabsBaseEvent: baseEvent}
	case "Complaint":
		event = &SocketLabsComplaintEvent{SocketLabsBaseEvent: baseEvent}
	case "Failed":
		event = &SocketLabsFailedEvent{SocketLabsBaseEvent: baseEvent}
	case "Delivered":
		event = &SocketLabsDeliveredEvent{SocketLabsBaseEvent: baseEvent}
	case "Queued":
		event = &SocketLabsQueuedEvent{SocketLabsBaseEvent: baseEvent}
	case "Deferred":
		event = &SocketLabsDeferredEvent{SocketLabsBaseEvent: baseEvent}
	default:
		log.Printf("Unknown event type: %s", baseEvent.Type)
		return
	}

	err = event.UnmarshalSocketLabsEvent(payload.Body, payload.Headers)
	if err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		// log.Printf("Raw payload: %s", string(payload.Body))
		return
	}

	// fmt.Printf("Event: %+v\n", event)
}
