package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

func ProcessPostmarkEvents(msg *sarama.ConsumerMessage) {
	var payload struct {
		Headers json.RawMessage `json:"headers"`
		Body    json.RawMessage `json:"body"`
	}
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

	err = event.UnmarshalPostmarkEvent(payload.Body)
	if err != nil {
		fmt.Printf("Failed to unmarshal event: %v\n", err)
		return
	}

	fmt.Printf("Event: %+v\n", event)
	saveToDatabase(event)
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
	UnmarshalPostmarkEvent(data []byte) error
}

type PostmarkDeliveryEvent struct {
	PostmarkEvent
}

func (p *PostmarkDeliveryEvent) UnmarshalPostmarkEvent(data []byte) error {
	if err := json.Unmarshal(data, p); err != nil {
		return err
	}
	p.Provider = "Postmark"
	return nil
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

func (p *PostmarkBounceEvent) UnmarshalPostmarkEvent(data []byte) error {
	if err := json.Unmarshal(data, p); err != nil {
		return err
	}
	p.Provider = "Postmark"
	return nil
}

type PostmarkSpamComplaintEvent struct {
	PostmarkEvent
	FromEmail   string    `json:"FromEmail"`
	BouncedAt   time.Time `json:"BouncedAt"`
	Subject     string    `json:"Subject"`
	MailboxHash string    `json:"MailboxHash"`
}

func (p *PostmarkSpamComplaintEvent) UnmarshalPostmarkEvent(data []byte) error {
	if err := json.Unmarshal(data, p); err != nil {
		return err
	}
	p.Provider = "Postmark"
	return nil
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

// Supporting structs for Open and Click events
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

func (p *PostmarkOpenEvent) UnmarshalPostmarkEvent(data []byte) error {
	if err := json.Unmarshal(data, p); err != nil {
		return err
	}
	p.Provider = "Postmark"
	return nil
}

type PostmarkClickEvent struct {
	Provider      string
	RecordType    string                 `json:"RecordType"`
	MessageStream string                 `json:"MessageStream"`
	Metadata      map[string]interface{} `json:"Metadata"`
	Recipient     string                 `json:"Recipient"`
	MessageID     string                 `json:"MessageID"`
	ReceivedAt    time.Time              `json:"ReceivedAt"`
	Platform      string                 `json:"Platform"`
	ClickLocation string                 `json:"ClickLocation"`
	OriginalLink  string                 `json:"OriginalLink"`
	Tag           string                 `json:"Tag"`
	UserAgent     string                 `json:"UserAgent"`
	OS            PostmarkOSInfo         `json:"OS"`
	Client        PostmarkClientInfo     `json:"Client"`
	Geo           PostmarkGeoInfo        `json:"Geo"`
}

func (p *PostmarkClickEvent) UnmarshalPostmarkEvent(data []byte) error {
	if err := json.Unmarshal(data, p); err != nil {
		return err
	}
	p.Provider = "Postmark"
	return nil
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

func (p *PostmarkSubscriptionChangeEvent) UnmarshalPostmarkEvent(data []byte) error {
	if err := json.Unmarshal(data, p); err != nil {
		return err
	}
	p.Provider = "Postmark"
	return nil
}
