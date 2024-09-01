package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"relay-go-consumer/database"

	"github.com/IBM/sarama"
	"github.com/lib/pq"
)

// WebhookPayload represents the entire webhook payload
type PostmarkWebhookPayload struct {
	Headers PostmarkWebhookHeaders `json:"headers"`
	Body    json.RawMessage        `json:"body"`
}

// WebhookHeaders represents the headers in the webhook payload
type PostmarkWebhookHeaders struct {
	AcceptEncoding      []string `json:"Accept-Encoding"`
	ContentLength       []string `json:"Content-Length"`
	ContentType         []string `json:"Content-Type"`
	Expect              []string `json:"Expect"`
	UserAgent           []string `json:"User-Agent"`
	XForwardedFor       []string `json:"X-Forwarded-For"`
	XForwardedHost      []string `json:"X-Forwarded-Host"`
	XForwardedProto     []string `json:"X-Forwarded-Proto"`
	XPmRetriesRemaining []string `json:"X-Pm-Retries-Remaining"`
	XPmWebhookEventId   []string `json:"X-Pm-Webhook-Event-Id"`
	XPmWebhookTraceId   []string `json:"X-Pm-Webhook-Trace-Id"`
}

func ProcessPostmarkEvents(msg *sarama.ConsumerMessage) {
	var payload PostmarkWebhookPayload
	err := json.Unmarshal(msg.Value, &payload)
	if err != nil {
		fmt.Printf("Failed to unmarshal message: %v\n", err)
		return
	}

	var baseEvent PostmarkEvent
	err = json.Unmarshal(payload.Body, &baseEvent)
	if err != nil {
		fmt.Printf("Failed to unmarshal base event: %v\n", err)
		return
	}

	var event PostmarkEventUnmarshaler

	switch baseEvent.RecordType {
	case "Delivery":
		event = &PostmarkDeliveryEvent{}
	case "Bounce":
		event = &PostmarkBounceEvent{}
	case "SpamComplaint":
		event = &PostmarkSpamComplaintEvent{}
	case "Open":
		event = &PostmarkOpenEvent{}
	case "Click":
		event = &PostmarkClickEvent{}
	case "SubscriptionChange":
		event = &PostmarkSubscriptionChangeEvent{}
	default:
		fmt.Printf("Unknown event type: %s\n", baseEvent.RecordType)
		return
	}

	err = event.UnmarshalPostmarkEvent(payload.Body, payload.Headers)
	if err != nil {
		fmt.Printf("Failed to unmarshal event: %v\n", err)
		return
	}

	fmt.Printf("Event: %+v\n", event)
}

type PostmarkEvent struct {
	RecordType  string                 `json:"RecordType"`
	ID          int64                  `json:"ID"`
	ServerID    int                    `json:"ServerID"`
	MessageID   string                 `json:"MessageID"`
	Recipient   string                 `json:"Recipient"`
	Tag         string                 `json:"Tag"`
	DeliveredAt time.Time              `json:"DeliveredAt"`
	Details     string                 `json:"Details"`
	Metadata    map[string]interface{} `json:"Metadata"`
	Provider    string
}

type PostmarkEventUnmarshaler interface {
	UnmarshalPostmarkEvent(data []byte, headers PostmarkWebhookHeaders) error
}

func (p *PostmarkEvent) UnmarshalPostmarkEvent(data []byte, headers PostmarkWebhookHeaders) error {
	if err := json.Unmarshal(data, p); err != nil {
		return err
	}
	p.Provider = "Postmark"

	return p.saveToDatabase(data, headers)
}

func (p *PostmarkEvent) saveToDatabase(eventData []byte, headers PostmarkWebhookHeaders) error {
	database.InitDB()
	db := database.GetDB()

	stmt, err := db.Prepare(`
		INSERT INTO postmarkeventwithheaders (
			record_type, id, server_id, message_id, recipient, tag, delivered_at, details, metadata, provider,
			event, event_data,
			accept_encoding, content_length, content_type, expect, user_agent, x_forwarded_for,
			x_forwarded_host, x_forwarded_proto, x_pm_retries_remaining, x_pm_webhook_event_id,
			x_pm_webhook_trace_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
		)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	metadataJSON, err := json.Marshal(p.Metadata)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(
		strings.ToLower(p.RecordType),
		p.ID,
		p.ServerID,
		p.MessageID,
		p.Recipient,
		p.Tag,
		p.DeliveredAt,
		p.Details,
		string(metadataJSON),
		p.Provider,
		strings.ToLower(p.RecordType),
		string(eventData),
		pq.Array(headers.AcceptEncoding),
		pq.Array(headers.ContentLength),
		pq.Array(headers.ContentType),
		pq.Array(headers.Expect),
		pq.Array(headers.UserAgent),
		pq.Array(headers.XForwardedFor),
		pq.Array(headers.XForwardedHost),
		pq.Array(headers.XForwardedProto),
		pq.Array(headers.XPmRetriesRemaining),
		pq.Array(headers.XPmWebhookEventId),
		pq.Array(headers.XPmWebhookTraceId),
	)

	return err
}

type PostmarkDeliveryEvent struct {
	PostmarkEvent
}

func (p *PostmarkDeliveryEvent) UnmarshalPostmarkEvent(data []byte, headers PostmarkWebhookHeaders) error {
	return p.PostmarkEvent.UnmarshalPostmarkEvent(data, headers)
}

type PostmarkBounceEvent struct {
	PostmarkEvent
	Type          string    `json:"Type"`
	TypeCode      int       `json:"TypeCode"`
	Name          string    `json:"Name"`
	Description   string    `json:"Description"`
	Email         string    `json:"Email"`
	BouncedAt     time.Time `json:"BouncedAt"`
	DumpAvailable bool      `json:"DumpAvailable"`
	Inactive      bool      `json:"Inactive"`
	CanActivate   bool      `json:"CanActivate"`
	Subject       string    `json:"Subject"`
	Content       string    `json:"Content"`
}

func (p *PostmarkBounceEvent) UnmarshalPostmarkEvent(data []byte, headers PostmarkWebhookHeaders) error {
	return p.PostmarkEvent.UnmarshalPostmarkEvent(data, headers)
}

type PostmarkSpamComplaintEvent struct {
	PostmarkEvent
	FromEmail   string    `json:"FromEmail"`
	BouncedAt   time.Time `json:"BouncedAt"`
	Subject     string    `json:"Subject"`
	MailboxHash string    `json:"MailboxHash"`
}

func (p *PostmarkSpamComplaintEvent) UnmarshalPostmarkEvent(data []byte, headers PostmarkWebhookHeaders) error {
	return p.PostmarkEvent.UnmarshalPostmarkEvent(data, headers)
}

type PostmarkOpenEvent struct {
	PostmarkEvent
	FirstOpen   bool               `json:"FirstOpen"`
	ReceivedAt  string             `json:"ReceivedAt"`
	Platform    string             `json:"Platform"`
	ReadSeconds int                `json:"ReadSeconds"`
	UserAgent   string             `json:"UserAgent"`
	OS          PostmarkOSInfo     `json:"OS"`
	Client      PostmarkClientInfo `json:"Client"`
	Geo         PostmarkGeoInfo    `json:"Geo"`
}

type PostmarkOSInfo struct {
	Name    string `json:"Name"`
	Family  string `json:"Family"`
	Company string `json:"Company"`
}

type PostmarkClientInfo struct {
	Name    string `json:"Name"`
	Family  string `json:"Family"`
	Company string `json:"Company"`
}

type PostmarkGeoInfo struct {
	IP             string `json:"IP"`
	City           string `json:"City"`
	Country        string `json:"Country"`
	CountryISOCode string `json:"CountryISOCode"`
	Region         string `json:"Region"`
	RegionISOCode  string `json:"RegionISOCode"`
	Zip            string `json:"Zip"`
	Coords         string `json:"Coords"`
}

func (p *PostmarkOpenEvent) UnmarshalPostmarkEvent(data []byte, headers PostmarkWebhookHeaders) error {
	return p.PostmarkEvent.UnmarshalPostmarkEvent(data, headers)
}

type PostmarkClickEvent struct {
	PostmarkEvent
	MessageStream string             `json:"MessageStream"`
	ReceivedAt    time.Time          `json:"ReceivedAt"`
	Platform      string             `json:"Platform"`
	ClickLocation string             `json:"ClickLocation"`
	OriginalLink  string             `json:"OriginalLink"`
	UserAgent     string             `json:"UserAgent"`
	OS            PostmarkOSInfo     `json:"OS"`
	Client        PostmarkClientInfo `json:"Client"`
	Geo           PostmarkGeoInfo    `json:"Geo"`
}

func (p *PostmarkClickEvent) UnmarshalPostmarkEvent(data []byte, headers PostmarkWebhookHeaders) error {
	return p.PostmarkEvent.UnmarshalPostmarkEvent(data, headers)
}

type PostmarkSubscriptionChangeEvent struct {
	PostmarkEvent
	SuppressSending   bool      `json:"SuppressSending"`
	SuppressionReason string    `json:"SuppressionReason"`
	ChangedAt         time.Time `json:"ChangedAt"`
	Source            string    `json:"Source"`
	SourceType        string    `json:"SourceType"`
	MessageStream     string    `json:"MessageStream"`
}

func (p *PostmarkSubscriptionChangeEvent) UnmarshalPostmarkEvent(data []byte, headers PostmarkWebhookHeaders) error {
	return p.PostmarkEvent.UnmarshalPostmarkEvent(data, headers)
}
