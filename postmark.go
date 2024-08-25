package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
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
	fmt.Printf("JSON data: %s\n", string(jsonData))

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
	requestDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		fmt.Printf("failed to dump HTTP request: %v\n", err)
	}
	fmt.Println("HTTP Request:")
	fmt.Println(string(requestDump))

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
			Tag:           "",                // Optional, set as needed
			HtmlBody:      emailMessage.Body, // Assuming Body is used as HtmlBody
			TextBody:      emailMessage.Body, // Assuming Body is used as TextBody
			ReplyTo:       "",                // Optional, set as needed
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
