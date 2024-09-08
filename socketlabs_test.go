package main

import (
	"fmt"
	"testing"

	"github.com/socketlabs/socketlabs-go/injectionapi/message"
	"github.com/stretchr/testify/assert"
)

func TestProcessContent(t *testing.T) {
	content := []Content{
		{Type: "text/plain", Value: "Hello {{name}}, your order -order_id- is ready."},
		{Type: "text/html", Value: "<p>Hello -name-, your order {{order_id}} is ready.</p>"},
	}
	substitutions := map[string]string{
		"name":     "John",
		"order_id": "12345",
	}
	sections := map[string]string{
		"greeting": "Welcome, {{name}}!",
	}

	processed := processContent(content, substitutions, sections)

	assert.Equal(t, "Hello John, your order 12345 is ready.", processed[0].Value)
	assert.Equal(t, "<p>Hello John, your order 12345 is ready.</p>", processed[1].Value)
}

func TestParseSectionsDynamic(t *testing.T) {
	sections := map[string]string{
		"-section1-":   "Content 1",
		"{{section2}}": "Content 2",
		"plainSection": "Content 3",
	}

	parsed := parseSectionsDynamic(sections)

	assert.Equal(t, "Content 1", parsed["section1"])
	assert.Equal(t, "Content 2", parsed["section2"])
	assert.Equal(t, "Content 3", parsed["plainSection"])
}

func TestGetContentByType(t *testing.T) {
	content := []Content{
		{Type: "text/plain", Value: "Plain text"},
		{Type: "text/html", Value: "<p>HTML content</p>"},
	}

	plainText := getContentByType(content, "text/plain")
	htmlContent := getContentByType(content, "text/html")

	assert.Equal(t, "Plain text", plainText)
	assert.Equal(t, "<p>HTML content</p>", htmlContent)
}

func TestPrepareSocketLabsMessages(t *testing.T) {
	emailMessage := EmailMessage{
		From: EmailAddress{Email: "sender@example.com", Name: "Sender"},
		Personalizations: []Personalization{
			{
				To:      EmailAddress{Email: "recipient1@example.com", Name: "Recipient1"},
				Subject: "Test Email 1",
				Substitutions: map[string]string{
					"name":     "John",
					"order_id": "12345",
				},
			},
			{
				To:      EmailAddress{Email: "recipient2@example.com", Name: "Recipient2"},
				Subject: "Test Email 2",
				Substitutions: map[string]string{
					"name":     "Jane",
					"order_id": "67890",
				},
			},
		},
		Content: []Content{
			{Type: "text/plain", Value: "Hello {{name}}, your order -order_id- is ready."},
			{Type: "text/html", Value: "<p>Hello -name-, your order {{order_id}} is ready.</p>"},
		},
		Sections: map[string]string{
			"greeting": "Welcome, {{name}}!",
		},
		Headers: map[string]string{
			"X-Custom-Header": "Custom Value",
		},
		Attachments: []Attachment{
			{
				Filename: "test.txt",
				Content:  "SGVsbG8gV29ybGQh", // Base64 encoded "Hello World!"
				Type:     "text/plain",
			},
		},
		Cc:  []string{"cc@example.com"},
		Bcc: []string{"bcc@example.com"},
		Credentials: Credentials{
			SocketLabsAPIKey: "test-api-key",
		},
	}

	preparedMessages := prepareSocketLabsMessages(emailMessage)

	assert.Equal(t, 2, len(preparedMessages), "Should have 2 prepared messages")

	for i, msg := range preparedMessages {
		assert.Equal(t, "sender@example.com", msg.From.EmailAddress)
		assert.Equal(t, "Sender", msg.From.FriendlyName)
		assert.Equal(t, fmt.Sprintf("recipient%d@example.com", i+1), msg.To[0].EmailAddress)
		assert.Equal(t, fmt.Sprintf("Test Email %d", i+1), msg.Subject)

		expectedName := "John"
		expectedOrderID := "12345"
		if i == 1 {
			expectedName = "Jane"
			expectedOrderID = "67890"
		}

		assert.Equal(t, fmt.Sprintf("Hello %s, your order %s is ready.", expectedName, expectedOrderID), msg.PlainTextBody)
		assert.Equal(t, fmt.Sprintf("<p>Hello %s, your order %s is ready.</p>", expectedName, expectedOrderID), msg.HtmlBody)

		assert.Equal(t, 1, len(msg.Cc))
		assert.Equal(t, "cc@example.com", msg.Cc[0].EmailAddress)

		assert.Equal(t, 1, len(msg.Bcc))
		assert.Equal(t, "bcc@example.com", msg.Bcc[0].EmailAddress)

		assert.Equal(t, 2, len(msg.CustomHeaders))
		assert.Contains(t, msg.CustomHeaders, message.CustomHeader{Name: "X-Custom-Header", Value: "Custom Value"})
		assert.Contains(t, msg.CustomHeaders, message.CustomHeader{Name: "X-xsMessageId"})

		assert.Equal(t, 1, len(msg.Attachments))
		assert.Equal(t, "test.txt", msg.Attachments[0].Name)
		assert.Equal(t, "text/plain", msg.Attachments[0].MimeType)
		assert.Equal(t, []byte("Hello World!"), msg.Attachments[0].Content)
	}
}
