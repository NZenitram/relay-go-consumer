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
type PostMarkDeliveryEvent struct {
	RecordType    string                 `json:"RecordType"`
	ServerID      int                    `json:"ServerID"`
	MessageStream string                 `json:"MessageStream"`
	MessageID     string                 `json:"MessageID"`
	Recipient     string                 `json:"Recipient"`
	Tag           string                 `json:"Tag"`
	DeliveredAt   string                 `json:"DeliveredAt"`
	Details       string                 `json:"Details"`
	Metadata      map[string]interface{} `json:"Metadata"`
}

// PostMarkBounceEvent struct
type PostMarkBounceEvent struct {
	ID            int                    `json:"ID"`
	Type          string                 `json:"Type"`
	RecordType    string                 `json:"RecordType"`
	TypeCode      int                    `json:"TypeCode"`
	Tag           string                 `json:"Tag"`
	MessageID     string                 `json:"MessageID"`
	Details       string                 `json:"Details"`
	Email         string                 `json:"Email"`
	From          string                 `json:"From"`
	BouncedAt     string                 `json:"BouncedAt"`
	Inactive      bool                   `json:"Inactive"`
	DumpAvailable bool                   `json:"DumpAvailable"`
	CanActivate   bool                   `json:"CanActivate"`
	Subject       string                 `json:"Subject"`
	ServerID      int                    `json:"ServerID"`
	MessageStream string                 `json:"MessageStream"`
	Content       string                 `json:"Content"`
	Name          string                 `json:"Name"`
	Description   string                 `json:"Description"`
	Metadata      map[string]interface{} `json:"Metadata"`
}

// PostMarkSpamComplaintEvent struct
type PostMarkSpamComplaintEvent struct {
	ID            int                    `json:"ID"`
	Type          string                 `json:"Type"`
	RecordType    string                 `json:"RecordType"`
	TypeCode      int                    `json:"TypeCode"`
	Tag           string                 `json:"Tag"`
	MessageID     string                 `json:"MessageID"`
	Details       string                 `json:"Details"`
	Email         string                 `json:"Email"`
	From          string                 `json:"From"`
	BouncedAt     string                 `json:"BouncedAt"`
	Inactive      bool                   `json:"Inactive"`
	DumpAvailable bool                   `json:"DumpAvailable"`
	CanActivate   bool                   `json:"CanActivate"`
	Subject       string                 `json:"Subject"`
	ServerID      int                    `json:"ServerID"`
	MessageStream string                 `json:"MessageStream"`
	Content       string                 `json:"Content"`
	Name          string                 `json:"Name"`
	Description   string                 `json:"Description"`
	Metadata      map[string]interface{} `json:"Metadata"`
}

// PostMarkOpenEvent struct
type PostMarkOpenEvent struct {
	RecordType    string                 `json:"RecordType"`
	MessageStream string                 `json:"MessageStream"`
	Metadata      map[string]interface{} `json:"Metadata"`
	FirstOpen     bool                   `json:"FirstOpen"`
	Recipient     string                 `json:"Recipient"`
	MessageID     string                 `json:"MessageID"`
	ReceivedAt    string                 `json:"ReceivedAt"`
	Platform      string                 `json:"Platform"`
	ReadSeconds   int                    `json:"ReadSeconds"`
	Tag           string                 `json:"Tag"`
	UserAgent     string                 `json:"UserAgent"`
	OS            PostMarkOSInfo         `json:"OS"`
	Client        PostMarkClientInfo     `json:"Client"`
	Geo           PostMarkGeoInfo        `json:"Geo"`
}

// PostMarkClickEvent struct
type PostMarkClickEvent struct {
	RecordType    string                 `json:"RecordType"`
	MessageStream string                 `json:"MessageStream"`
	Metadata      map[string]interface{} `json:"Metadata"`
	Recipient     string                 `json:"Recipient"`
	MessageID     string                 `json:"MessageID"`
	ReceivedAt    string                 `json:"ReceivedAt"`
	Platform      string                 `json:"Platform"`
	ClickLocation string                 `json:"ClickLocation"`
	OriginalLink  string                 `json:"OriginalLink"`
	Tag           string                 `json:"Tag"`
	UserAgent     string                 `json:"UserAgent"`
	OS            PostMarkOSInfo         `json:"OS"`
	Client        PostMarkClientInfo     `json:"Client"`
	Geo           PostMarkGeoInfo        `json:"Geo"`
}

// PostMarkSubscriptionChangeEvent struct
type PostMarkSubscriptionChangeEvent struct {
	RecordType        string                 `json:"RecordType"`
	MessageID         string                 `json:"MessageID"`
	ServerID          int                    `json:"ServerID"`
	MessageStream     string                 `json:"MessageStream"`
	ChangedAt         string                 `json:"ChangedAt"`
	Recipient         string                 `json:"Recipient"`
	Origin            string                 `json:"Origin"`
	SuppressSending   bool                   `json:"SuppressSending"`
	SuppressionReason string                 `json:"SuppressionReason"`
	Tag               string                 `json:"Tag"`
	Metadata          map[string]interface{} `json:"Metadata"`
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
