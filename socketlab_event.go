package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

// SocketLabsEventUnmarshaler interface
type SocketLabsEventUnmarshaler interface {
	UnmarshalSocketLabsEvent(data []byte) error
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

// SocketLabsTrackingEvent struct for Click events
type SocketLabsTrackingEvent struct {
	SocketLabsBaseEvent
	TrackingType int            `json:"TrackingType"`
	ClientIp     string         `json:"ClientIp"`
	Url          string         `json:"Url"`
	UserAgent    string         `json:"UserAgent"`
	Data         SocketLabsData `json:"Data"`
}

// SocketLabsComplaintEvent struct for Complaint events
type SocketLabsComplaintEvent struct {
	SocketLabsBaseEvent
	UserAgent string `json:"UserAgent"`
	From      string `json:"From"`
	To        string `json:"To"`
	Length    int    `json:"Length"`
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

// SocketLabsDeliveredEvent struct for Delivered events
type SocketLabsDeliveredEvent struct {
	SocketLabsBaseEvent
	Response  string         `json:"Response"`
	LocalIp   string         `json:"LocalIp"`
	RemoteMta string         `json:"RemoteMta"`
	Data      SocketLabsData `json:"Data"`
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

// SocketLabsDeferredEvent struct for Deferred events
type SocketLabsDeferredEvent struct {
	SocketLabsBaseEvent
	FromAddress  string         `json:"FromAddress"`
	DeferralCode int            `json:"DeferralCode"`
	Reason       string         `json:"Reason"`
	Data         SocketLabsData `json:"Data"`
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

// UnmarshalSocketLabsEvent methods for each event type
func (e *SocketLabsTrackingEvent) UnmarshalSocketLabsEvent(data []byte) error {
	return json.Unmarshal(data, e)
}

func (e *SocketLabsComplaintEvent) UnmarshalSocketLabsEvent(data []byte) error {
	return json.Unmarshal(data, e)
}

func (e *SocketLabsFailedEvent) UnmarshalSocketLabsEvent(data []byte) error {
	return json.Unmarshal(data, e)
}

func (e *SocketLabsDeliveredEvent) UnmarshalSocketLabsEvent(data []byte) error {
	return json.Unmarshal(data, e)
}

func (e *SocketLabsQueuedEvent) UnmarshalSocketLabsEvent(data []byte) error {
	return json.Unmarshal(data, e)
}

func (e *SocketLabsDeferredEvent) UnmarshalSocketLabsEvent(data []byte) error {
	return json.Unmarshal(data, e)
}

func (e *SocketLabsTrackingEvent) GetTrackingTypeString() string {
	if typeName, ok := trackingTypeMap[e.TrackingType]; ok {
		e.Type = typeName
	}
	return "Unknown"
}

func ProcessSocketLabsEvents(msg *sarama.ConsumerMessage) {
	var payload struct {
		Headers json.RawMessage `json:"headers"`
		Body    json.RawMessage `json:"body"`
	}
	err := json.Unmarshal(msg.Value, &payload)
	if err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	var baseEvent SocketLabsBaseEvent
	err = json.Unmarshal(payload.Body, &baseEvent)
	if err != nil {
		log.Printf("Failed to unmarshal base event: %v", err)
		log.Printf("Raw payload: %s", string(payload.Body))
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

	err = event.UnmarshalSocketLabsEvent(payload.Body)
	if err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		log.Printf("Raw payload: %s", string(payload.Body))
		return
	}

	// Handle TrackingType mapping for SocketLabsTrackingEvent
	if trackingEvent, ok := event.(*SocketLabsTrackingEvent); ok {
		trackingTypeString := trackingEvent.GetTrackingTypeString()
		fmt.Printf("Event Type: %s, Tracking Type: %s (Code: %d)\n",
			trackingEvent.Type, trackingTypeString, trackingEvent.TrackingType)
	}

	fmt.Printf("Event: %+v\n", event)
	saveToDatabase(event)
}
