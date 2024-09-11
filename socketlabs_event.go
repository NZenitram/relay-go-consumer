package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

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
	err = saveStandardizedEvent(standardizedEvent)
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
func generateMessageID(secretKey string, serverID int) string {
	// Use the current timestamp to ensure uniqueness
	timestamp := time.Now().UTC().UnixNano()

	// Create a unique identifier based on SecretKey, ServerID, and timestamp
	uniqueString := fmt.Sprintf("%s:%d:%d", secretKey, serverID, timestamp)

	// Use base64 encoding to create a URL-safe string
	encoded := base64.URLEncoding.EncodeToString([]byte(uniqueString))

	// Trim any padding characters
	return strings.TrimRight(encoded, "=")
}
