package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendEmailWithSendGrid(emailMessage EmailMessage) {
	apiKey := emailMessage.Credentials.SendgridAPIKey
	client := sendgrid.NewSendClient(apiKey)
	// Parse sections dynamically
	parsedSections := parseSectionsDynamic(emailMessage.Sections)
	// Transform substitutions
	transformedPersonalizations := transformSubstitutionsDynamic(emailMessage.Personalizations, parsedSections)

	for _, p := range transformedPersonalizations {
		message := mail.NewV3Mail()

		from := mail.NewEmail(emailMessage.From.Name, emailMessage.From.Email)
		message.SetFrom(from)

		// Set subject (personalized or default)
		subject := p.Subject
		if subject == "" {
			subject = emailMessage.Subject
		}
		message.Subject = subject

		// Add recipient
		to := mail.NewEmail(p.To.Name, p.To.Email)
		personalization := mail.NewPersonalization()
		personalization.AddTos(to)

		// Add CC and BCC recipients
		for _, cc := range emailMessage.Cc {
			personalization.AddCCs(mail.NewEmail("", cc))
		}
		for _, bcc := range emailMessage.Bcc {
			personalization.AddBCCs(mail.NewEmail("", bcc))
		}

		// Process content with substitutions
		for _, content := range emailMessage.Content {
			processedValue := content.Value
			for key, value := range p.Substitutions {
				processedValue = strings.ReplaceAll(processedValue, key, value)
			}
			for key, value := range emailMessage.Sections {
				processedValue = strings.ReplaceAll(processedValue, key, value)
			}
			message.AddContent(mail.NewContent(content.Type, processedValue))
		}

		message.Sections = emailMessage.Sections

		// Add transformed substitutions
		personalization.Substitutions = p.Substitutions

		message.AddPersonalizations(personalization)

		// Add attachments
		for _, attachment := range emailMessage.Attachments {
			content, err := base64.StdEncoding.DecodeString(attachment.Content)
			if err != nil {
				log.Printf("Failed to decode attachment content: %v", err)
				continue
			}

			message.AddAttachment(&mail.Attachment{
				Content:     string(content),
				ContentID:   attachment.ContentID,
				Type:        attachment.Type,
				Filename:    attachment.Filename,
				Name:        attachment.Name,
				Disposition: attachment.Disposition,
			})
		}

		// Add custom headers
		message.Headers = emailMessage.Headers

		// Add categories
		message.Categories = emailMessage.Categories

		// printMessageStructure(message)
		// Send the emails
		response, err := client.Send(message)
		if err != nil {
			SendGridErrorHandler(response, err, p.To.Email)
		}
	}
}

func printMessageStructure(message *mail.SGMailV3) {
	jsonData, err := json.MarshalIndent(message, "", "    ")
	if err != nil {
		log.Fatalf("Error marshaling message to JSON: %v", err)
	}
	fmt.Println(string(jsonData))
}

func transformSubstitutionsDynamic(personalizations []Personalization, sections map[string]string) []Personalization {
	for i, personalization := range personalizations {
		newSubstitutions := make(map[string]string)

		for key, value := range personalization.Substitutions {
			// Create both hyphen and handlebars versions of the key
			hyphenKey := fmt.Sprintf("-%s-", key)
			handlebarKey := fmt.Sprintf("{{%s}}", key)

			// Add both versions to the new substitutions
			newSubstitutions[hyphenKey] = value
			newSubstitutions[handlebarKey] = value

			// Check if the value corresponds to a section key
			hyphenSectionKey := fmt.Sprintf("-%s-", value)
			handlebarSectionKey := fmt.Sprintf("{{%s}}", value)
			if _, exists := sections[hyphenSectionKey]; exists {
				newSubstitutions[hyphenKey] = hyphenSectionKey
				// newSubstitutions[handlebarKey] = hyphenSectionKey
			} else if _, exists := sections[handlebarSectionKey]; exists {
				newSubstitutions[hyphenKey] = handlebarSectionKey
				newSubstitutions[handlebarKey] = handlebarSectionKey
			}
		}

		personalizations[i].Substitutions = newSubstitutions
	}

	return personalizations
}

func parseSectionsDynamic(sections map[string]string) map[string]string {
	newSections := make(map[string]string)
	for key, value := range sections {
		// Keep the original key
		newSections[key] = value

		// Add handlebars version if it's a hyphen-wrapped key
		if strings.HasPrefix(key, "-") && strings.HasSuffix(key, "-") {
			handlebarKey := fmt.Sprintf("{{%s}}", strings.Trim(key, "-"))
			newSections[handlebarKey] = value
		}

		// Add hyphen-wrapped version if it's a handlebars key
		if strings.HasPrefix(key, "{{") && strings.HasSuffix(key, "}}") {
			hyphenKey := fmt.Sprintf("-%s-", strings.Trim(key, "{}"))
			newSections[hyphenKey] = value
		}
	}
	return newSections
}
