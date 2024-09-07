package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	sp "github.com/SparkPost/gosparkpost"
)

func SendEmailWithSparkPost(emailMessage EmailMessage) {
	apiKey := emailMessage.Credentials.SparkpostAPIKey
	if apiKey == "" {
		log.Fatal("Missing SparkPost API key in credentials")
	}

	cfg := &sp.Config{
		BaseUrl:    "https://api.sparkpost.com",
		ApiKey:     apiKey,
		ApiVersion: 1,
	}
	var client sp.Client
	err := client.Init(cfg)
	if err != nil {
		log.Fatalf("SparkPost client init failed: %s\n", err)
	}

	globalSubstitutionData := make(map[string]string)
	for key, value := range emailMessage.Sections {
		// Remove hyphens from keys
		cleanKey := strings.Trim(key, "-")

		// Replace leading and trailing '-' with '{{' and '}}' in the value
		re := regexp.MustCompile(`-(\w+)-`)
		updatedValue := re.ReplaceAllString(value, "{{$1}}")

		globalSubstitutionData[cleanKey] = updatedValue
	}

	// Prepare recipients and substitution data
	recipients := make([]sp.Recipient, len(emailMessage.Personalizations))
	for i, p := range emailMessage.Personalizations {

		recipientSubstitutions := make(map[string]interface{})

		for key, value := range p.Substitutions {
			// Check if the value matches a key in globalSubstitutionData
			if sectionContent, exists := globalSubstitutionData[value]; exists {
				// Process placeholders in the section content using the recipient's substitutions
				processedContent := processPlaceholders(sectionContent, p.Substitutions)
				recipientSubstitutions[key] = processedContent
			} else {
				// If no match found, use the original value
				recipientSubstitutions[key] = value
			}
		}

		// Process placeholders in the substitution values
		// for key, value := range recipientSubstitutions {
		// 	if strValue, ok := value.(string); ok {
		// 		recipientSubstitutions[key] = processPlaceholders(strValue, p.Substitutions)
		// 	}
		// }

		recipients[i] = sp.Recipient{
			Address: sp.Address{
				Email: p.To.Email,
				Name:  p.To.Name,
			},
			SubstitutionData: recipientSubstitutions,
		}
	}

	// Add CC and BCC recipients
	for _, cc := range emailMessage.Cc {
		recipients = append(recipients, sp.Recipient{Address: sp.Address{Email: cc}})
	}
	for _, bcc := range emailMessage.Bcc {
		recipients = append(recipients, sp.Recipient{Address: sp.Address{Email: bcc}})
	}

	// Prepare attachments
	attachments := make([]sp.Attachment, len(emailMessage.Attachments))
	for i, att := range emailMessage.Attachments {
		content, err := base64.StdEncoding.DecodeString(att.Content)
		if err != nil {
			log.Printf("Failed to decode attachment content: %v", err)
			continue
		}
		attachments[i] = sp.Attachment{
			Filename: att.Filename,
			MIMEType: att.Type,
			B64Data:  string(content),
		}
	}

	// Prepare content
	htmlContent := ""
	textContent := ""
	for _, content := range emailMessage.Content {
		// Use the first personalization's substitutions for processing placeholders
		re := regexp.MustCompile(`-(\w+)-`)
		updatedValue := re.ReplaceAllString(content.Value, "{{$1}}")
		if content.Type == "text/html" {
			htmlContent = updatedValue
		} else if content.Type == "text/plain" {
			textContent = updatedValue
		}
	}

	// Create a Transmission
	tx := &sp.Transmission{
		Recipients: recipients,
		Content: sp.Content{
			From:        sp.Address{Email: emailMessage.From.Email, Name: emailMessage.From.Name},
			Subject:     emailMessage.Subject,
			HTML:        htmlContent,
			Text:        textContent,
			Headers:     emailMessage.Headers,
			Attachments: attachments,
		},
		SubstitutionData: globalSubstitutionData,
	}

	printSPMessageStructure(tx)

	// Send the email
	id, res, err := client.Send(tx)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return
	}

	log.Printf("Email sent successfully. ID: %s, Response: %v", id, res)
}

func processPlaceholders(content string, substitutions map[string]string) string {
	re := regexp.MustCompile(`\{\{(.*?)\}\}`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		placeholder := strings.Trim(match, "{}")
		if substitutionValue, exists := substitutions[placeholder]; exists {
			return substitutionValue
		}
		return match
	})
}

func printSPMessageStructure(message *sp.Transmission) {
	jsonData, err := json.MarshalIndent(message, "", "    ")
	if err != nil {
		log.Fatalf("Error marshaling message to JSON: %v", err)
	}
	fmt.Println(string(jsonData))
}
