package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

type PostmarkWebhookPayload struct {
	Headers PostmarkWebhookHeaders `json:"headers"`
	Body    json.RawMessage        `json:"body"`
}

type PostmarkWebhookHeaders struct {
	AcceptEncoding      []string `json:"Accept-Encoding"`
	Authorization       []string `json:"Authorization"`
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

	standardizedEvent := standardizePostmarkEvent(baseEvent)
	err = saveStandardizedEvent(standardizedEvent)
	if err != nil {
		fmt.Printf("Failed to save standardized event: %v\n", err)
	}
}

type PostmarkEvent struct {
	RecordType  string    `json:"RecordType"`
	ServerID    int       `json:"ServerID"`
	MessageID   string    `json:"MessageID"`
	Recipient   string    `json:"Recipient"`
	Tag         string    `json:"Tag"`
	DeliveredAt time.Time `json:"DeliveredAt"`
	Details     string    `json:"Details"`
	Type        string    `json:"Type"`
	TypeCode    int       `json:"TypeCode"`
	BouncedAt   time.Time `json:"BouncedAt"`
	BounceEmail string    `json:"Email"`
	ReceivedAt  time.Time `json:"ReceivedAt"`
}

func standardizePostmarkEvent(event PostmarkEvent) StandardizedEvent {
	standardEvent := StandardizedEvent{
		MessageID:     event.MessageID,
		Provider:      "postmark",
		Processed:     true,
		ProcessedTime: time.Now().UTC().Unix(),
	}

	switch event.RecordType {
	case "Delivery":
		standardEvent.Delivered = true
		deliveredTime := event.DeliveredAt.Unix()
		standardEvent.DeliveredTime = &deliveredTime
	case "Bounce":
		standardEvent.Bounce = true
		standardEvent.BounceType = event.Type
		bounceTime := event.BouncedAt.Unix()
		standardEvent.BounceTime = &bounceTime
		// For hard bounces, we might want to mark it as dropped as well
		if event.Type == "HardBounce" {
			standardEvent.Dropped = true
			standardEvent.DroppedTime = &bounceTime
			standardEvent.DroppedReason = event.Details
		}
	case "Open":
		standardEvent.Open = true
		standardEvent.UniqueOpen = true
		openTime := event.ReceivedAt.Unix()
		standardEvent.UniqueOpenTime = &openTime
		standardEvent.OpenCount = 1
	}

	return standardEvent
}
