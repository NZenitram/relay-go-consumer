package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	apiURL := "https://api.postmarkapp.com/email"

	// Strip credentials from the email message
	emailMessage.Credentials = Credentials{}

	postmarkMessages := mapEmailMessageToPostmark(emailMessage)

	for _, msg := range postmarkMessages {
		jsonData, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal email message: %v", err)
		}

		// Create a new HTTP request
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %v", err)
		}

		// Set default headers
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Postmark-Server-Token", serverToken)

		// Send the request
		sendEmail(req)
	}
	return nil
}

func sendEmail(req *http.Request) error {
	// Dump the request to a byte slice
	// dump, err := httputil.DumpRequestOut(req, true)
	// if err != nil {
	// 	return fmt.Errorf("failed to dump request: %v", err)
	// }

	// // Print the dumped request
	// log.Printf("Request:\n%s", string(dump))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}

	return HandlePostmarkResponse(resp, err)
}

func mapEmailMessageToPostmark(emailMessage EmailMessage) []PostMarkMessage {
	headers := make([]CustomHeader, 0, len(emailMessage.Headers))
	for key, value := range emailMessage.Headers {
		headers = append(headers, CustomHeader{Name: key, Value: value})
	}

	parsedSections := parseSectionsDynamicPostMark(emailMessage.Sections)
	var postMarkMessages []PostMarkMessage

	for _, personalization := range emailMessage.Personalizations {
		// Process content for each personalization
		processedContent := processContent(emailMessage.Content, personalization.Substitutions, parsedSections)

		postMarkMessage := PostMarkMessage{
			From:          emailMessage.From.Email,
			To:            personalization.To.Email,
			Cc:            strings.Join(emailMessage.Cc, ", "),
			Bcc:           strings.Join(emailMessage.Bcc, ", "),
			Subject:       personalization.Subject,
			Tag:           "",
			HtmlBody:      getContentByType(processedContent, "text/html"),
			TextBody:      getContentByType(processedContent, "text/plain"),
			ReplyTo:       "",
			Headers:       headers,
			Attachments:   convertAttachments(emailMessage.Attachments),
			TrackOpens:    true,
			TrackLinks:    "HtmlOnly",
			MessageStream: "outbound",
		}

		postMarkMessages = append(postMarkMessages, postMarkMessage)
	}

	return postMarkMessages
}

func processContent(content []Content, substitutions map[string]string, sections map[string]string) []Content {
	processedContent := make([]Content, len(content))
	for i, item := range content {
		value := item.Value

		// First pass: Replace section placeholders
		for sectionKey, sectionContent := range sections {
			handlebarPlaceholder := fmt.Sprintf("{{%s}}", sectionKey)
			hyphenPlaceholder := fmt.Sprintf("-%s-", sectionKey)
			value = strings.ReplaceAll(value, handlebarPlaceholder, sectionContent)
			value = strings.ReplaceAll(value, hyphenPlaceholder, sectionContent)
		}

		// Second pass: Replace substitutions
		for subKey, subValue := range substitutions {
			handlebarPlaceholder := fmt.Sprintf("{{%s}}", subKey)
			hyphenPlaceholder := fmt.Sprintf("-%s-", subKey)

			replacementValue := subValue
			if sectionContent, exists := sections[subValue]; exists {
				replacementValue = sectionContent
			}

			value = strings.ReplaceAll(value, handlebarPlaceholder, replacementValue)
			value = strings.ReplaceAll(value, hyphenPlaceholder, replacementValue)
		}

		// Third pass: Replace any remaining placeholders in the content
		for subKey, subValue := range substitutions {
			handlebarPlaceholder := fmt.Sprintf("{{%s}}", subKey)
			hyphenPlaceholder := fmt.Sprintf("-%s-", subKey)
			value = strings.ReplaceAll(value, handlebarPlaceholder, subValue)
			value = strings.ReplaceAll(value, hyphenPlaceholder, subValue)
		}

		processedContent[i] = Content{Type: item.Type, Value: value}
	}
	return processedContent
}

func getContentByType(content []Content, contentType string) string {
	for _, item := range content {
		if item.Type == contentType {
			return item.Value
		}
	}
	return ""
}

func convertAttachments(attachments []Attachment) []Attachment {
	postmarkAttachments := make([]Attachment, len(attachments))
	for i, att := range attachments {
		postmarkAttachments[i] = Attachment{
			Content:     att.Content,
			ContentID:   att.ContentID,
			Disposition: att.Disposition,
			Filename:    att.Filename,
			Name:        att.Name,
			ContentType: att.Type, // Use the 'type' field as 'ContentType'
		}
	}
	return postmarkAttachments
}

func parseSectionsDynamicPostMark(sections map[string]string) map[string]string {
	parsedSections := make(map[string]string)
	for key, value := range sections {
		// Remove any leading and trailing hyphens or curly braces
		cleanKey := strings.Trim(key, "-{}")
		parsedSections[cleanKey] = value
	}
	return parsedSections
}
