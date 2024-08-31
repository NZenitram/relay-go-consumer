package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type PostMarkMessage struct {
	From          string            `json:"From"`
	To            string            `json:"To"`
	Cc            string            `json:"Cc"`
	Bcc           string            `json:"Bcc"`
	Subject       string            `json:"Subject"`
	Tag           string            `json:"Tag"`
	HtmlBody      string            `json:"HtmlBody"`
	TextBody      string            `json:"TextBody"`
	ReplyTo       string            `json:"ReplyTo"`
	Metadata      map[string]string `json:"Metadata"`
	Headers       []CustomHeader    `json:"Headers"`
	Attachments   []Attachment      `json:"Attachments"`
	TrackOpens    bool              `json:"TrackOpens"`
	TrackLinks    string            `json:"TrackLinks"`
	MessageStream string            `json:"MessageStream"`
}

type CustomHeader struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

func SendEmailWithPostmark(emailMessage EmailMessage) error {
	// Extract credentials from the email message
	serverToken := emailMessage.Credentials.PostmarkServerToken
	apiURL := emailMessage.Credentials.PostmarkAPIURL

	// Strip credentials from the email message
	emailMessage.Credentials = Credentials{}

	postmarkMessage := mapEmailMessageToPostmark(emailMessage)

	// Marshal the email message to JSON
	jsonData, err := json.Marshal(postmarkMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal email message: %v", err)
	}
	// fmt.Printf("JSON data: %s\n", string(jsonData))

	// Create a new HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set default headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", serverToken)

	// Print the HTTP request to the console
	// requestDump, err := httputil.DumpRequestOut(req, true)
	// if err != nil {
	// 	fmt.Printf("failed to dump HTTP request: %v\n", err)
	// }
	// fmt.Println("HTTP Request:")
	// fmt.Println(string(requestDump))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Check for success status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send email, status: %s, response: %s", resp.Status, string(body))
	}

	log.Printf("Email sent successfully. Response: %s", string(body))
	return nil
}

func mapEmailMessageToPostmark(emailMessage EmailMessage) PostMarkMessage {
	// Convert headers map to a slice of CustomHeader
	headers := make([]CustomHeader, 0, len(emailMessage.Headers))
	for key, value := range emailMessage.Headers {
		headers = append(headers, CustomHeader{Name: key, Value: value})
	}

	// Create a PostMarkMessage for each recipient in the To field
	var postMarkMessage PostMarkMessage
	for _, addr := range emailMessage.To {
		postMarkMessage = PostMarkMessage{
			From:          emailMessage.From.Email,
			To:            addr.Email,
			Cc:            strings.Join(emailMessage.Cc, ", "),
			Bcc:           strings.Join(emailMessage.Bcc, ", "),
			Subject:       emailMessage.Subject,
			Tag:           "",                    // Optional, set as needed
			HtmlBody:      emailMessage.HtmlBody, // Assuming Body is used as HtmlBody
			TextBody:      emailMessage.TextBody, // Assuming Body is used as TextBody
			ReplyTo:       "",                    // Optional, set as needed
			Metadata:      emailMessage.AdditionalData,
			Headers:       headers,
			Attachments:   emailMessage.Attachments,
			TrackOpens:    true,       // or false, set as needed
			TrackLinks:    "HtmlOnly", // or other options, set as needed
			MessageStream: "outbound", // or other options, set as needed
		}
	}
	return postMarkMessage
}

// PostMarkDeliveryEvent struct

// Base struct for common fields
type PostMarkBaseEvent struct {
	Provider      string                 `json:"Provider"`
	RecordType    string                 `json:"RecordType"`
	MessageStream string                 `json:"MessageStream"`
	MessageID     string                 `json:"MessageID"`
	Recipient     string                 `json:"Recipient,omitempty"`
	Tag           string                 `json:"Tag,omitempty"`
	Metadata      map[string]interface{} `json:"Metadata,omitempty"`
}

// PostMarkDeliveryEvent struct
type PostMarkDeliveryEvent struct {
	PostMarkBaseEvent
	ServerID    int    `json:"ServerID"`
	DeliveredAt string `json:"DeliveredAt"`
	Details     string `json:"Details"`
}

// PostMarkBounceEvent struct
type PostMarkBounceEvent struct {
	PostMarkBaseEvent
	ID            int    `json:"ID"`
	Type          string `json:"Type"`
	TypeCode      int    `json:"TypeCode"`
	Details       string `json:"Details"`
	Email         string `json:"Email"`
	From          string `json:"From"`
	BouncedAt     string `json:"BouncedAt"`
	Inactive      bool   `json:"Inactive"`
	DumpAvailable bool   `json:"DumpAvailable"`
	CanActivate   bool   `json:"CanActivate"`
	Subject       string `json:"Subject"`
	ServerID      int    `json:"ServerID"`
	Content       string `json:"Content"`
	Name          string `json:"Name"`
	Description   string `json:"Description"`
}

// PostMarkSpamComplaintEvent struct
type PostMarkSpamComplaintEvent struct {
	PostMarkBounceEvent
}

// PostMarkOpenEvent struct
type PostMarkOpenEvent struct {
	PostMarkBaseEvent
	FirstOpen   bool               `json:"FirstOpen"`
	ReceivedAt  string             `json:"ReceivedAt"`
	Platform    string             `json:"Platform"`
	ReadSeconds int                `json:"ReadSeconds"`
	UserAgent   string             `json:"UserAgent"`
	OS          PostMarkOSInfo     `json:"OS"`
	Client      PostMarkClientInfo `json:"Client"`
	Geo         PostMarkGeoInfo    `json:"Geo"`
}

// PostMarkClickEvent struct
type PostMarkClickEvent struct {
	PostMarkOpenEvent
	ClickLocation string `json:"ClickLocation"`
	OriginalLink  string `json:"OriginalLink"`
}

// PostMarkSubscriptionChangeEvent struct
type PostMarkSubscriptionChangeEvent struct {
	PostMarkBaseEvent
	ServerID          int    `json:"ServerID"`
	ChangedAt         string `json:"ChangedAt"`
	Origin            string `json:"Origin"`
	SuppressSending   bool   `json:"SuppressSending"`
	SuppressionReason string `json:"SuppressionReason"`
}

// Supporting structs for Open and Click events
type PostMarkOSInfo struct {
	Name    string `json:"Name"`
	Family  string `json:"Family"`
	Company string `json:"Company"`
}

type PostMarkClientInfo struct {
	Name    string `json:"Name"`
	Family  string `json:"Family"`
	Company string `json:"Company"`
}

type PostMarkGeoInfo struct {
	IP             string `json:"IP"`
	City           string `json:"City"`
	Country        string `json:"Country"`
	CountryISOCode string `json:"CountryISOCode"`
	Region         string `json:"Region"`
	RegionISOCode  string `json:"RegionISOCode"`
	Zip            string `json:"Zip"`
	Coords         string `json:"Coords"`
}
