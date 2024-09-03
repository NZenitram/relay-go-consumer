package main

import (
	"encoding/json"
	"fmt"
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
