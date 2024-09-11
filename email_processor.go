package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"relay-go-consumer/database"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

type KafkaMessage struct {
	MessageID string              `json:"MessageID"`
	UserID    int                 `json:"UserID"`
	Headers   map[string][]string `json:"headers"`
	Body      EmailMessage        `json:"body"`
}

func ProcessEmailMessages(msg *sarama.ConsumerMessage) {
	var kafkaMessage KafkaMessage
	err := json.Unmarshal(msg.Value, &kafkaMessage)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// Fetch ESP credentials from the database
	credentials, credentialsErr := fetchESPCredentials(kafkaMessage.UserID)
	if credentialsErr != nil {
		fmt.Printf("failed to fetch ESP credentials: %v", credentialsErr)
	}

	// Initialize weights for active providers

	database.InitDB()
	db := database.GetDB()

	emailMessage := kafkaMessage.Body
	emailMessage.Credentials = credentials
	// Calculate weights based on event data
	weights, err := calculateWeights(db, kafkaMessage.UserID, credentials)
	if err != nil {
		fmt.Printf("failed to calculate weights: %v", err)
		return
	}

	// If there are no personalizations, create one for each recipient
	if len(emailMessage.Personalizations) == 0 {
		for _, recipient := range emailMessage.To {
			emailMessage.Personalizations = append(emailMessage.Personalizations, Personalization{
				To:            recipient,
				Subject:       emailMessage.Subject,
				Substitutions: make(map[string]string),
			})
		}
	}

	senderGroups := make(map[string][]Personalization)
	for _, p := range emailMessage.Personalizations {
		sender := SelectSender(weights)
		senderGroups[sender] = append(senderGroups[sender], p)
	}

	// Send emails using each selected sender
	for sender, personalizations := range senderGroups {
		groupMessage := emailMessage
		groupMessage.Personalizations = personalizations

		switch sender {
		case "SendGrid":
			SendEmailWithSendGrid(groupMessage)
		case "SocketLabs":
			SendEmailWithSocketLabs(groupMessage)
		case "Postmark":
			SendEmailWithPostmark(groupMessage)
		case "SparkPost":
			SendEmailWithSparkPost(groupMessage)
		default:
			log.Printf("No valid credentials found for sender: %s", sender)
		}
	}
}

func SelectSender(weights map[string]int) string {
	totalWeight := 0
	for _, weight := range weights {
		totalWeight += weight
	}

	if totalWeight == 0 {
		return ""
	}

	randomValue := rand.Intn(totalWeight)
	cumulativeWeight := 0

	for provider, weight := range weights {
		cumulativeWeight += weight
		if randomValue < cumulativeWeight {
			return provider
		}
	}

	return "" // This should never happen if weights are calculated correctly
}

func isValidProvider(provider string, credentials Credentials) bool {
	switch provider {
	case "SocketLabs":
		return credentials.SocketLabsServerID != "" && credentials.SocketLabsAPIKey != ""
	case "Postmark":
		return credentials.PostmarkServerToken != ""
	case "SendGrid":
		return credentials.SendgridAPIKey != ""
	case "SparkPost":
		return credentials.SparkpostAPIKey != ""
	default:
		return false
	}
}

// Custom unmarshaling logic for EmailAddress
func (e *EmailAddress) UnmarshalJSON(data []byte) error {
	// Attempt to unmarshal as a simple string
	var emailString string
	if err := json.Unmarshal(data, &emailString); err == nil {
		// Updated regex pattern to capture Friendly Name and Email Address
		r := regexp.MustCompile(`(?i)(?:"?([^"<]*)"?\s*<([^>]+)>|([^<>\s]+@[^<>\s]+))`)
		matches := r.FindStringSubmatch(emailString)
		if len(matches) > 0 {
			e.Name = strings.TrimSpace(matches[1])
			if matches[2] != "" {
				e.Email = strings.TrimSpace(matches[2])
			} else {
				e.Email = strings.TrimSpace(matches[3])
			}
		} else {
			e.Email = emailString
			e.Name = "" // No name available
		}
		return nil
	}

	// Attempt to unmarshal as an object with email and name
	var alias struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	e.Email = alias.Email
	e.Name = alias.Name
	return nil
}

type Attachment struct {
	Content     string `json:"content"`
	ContentID   string `json:"content_id"`
	Disposition string `json:"disposition"`
	Filename    string `json:"filename"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	ContentType string `json:"ContentType,omitempty"`
}

type EmailMessage struct {
	From             EmailAddress
	To               []EmailAddress
	Cc               []string
	Bcc              []string
	Subject          string
	TextBody         string
	HtmlBody         string
	Content          []Content
	Attachments      []Attachment
	Headers          map[string]string
	Data             map[string]interface{}
	Credentials      Credentials
	Personalizations []Personalization
	Sections         map[string]string
	Categories       []string
}
type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Personalization struct {
	To            EmailAddress
	Subject       string
	Substitutions map[string]string
}

type EmailAddress struct {
	Name  string
	Email string
}

type Credentials struct {
	SocketLabsServerID  string `json:"SocketLabsServerID"`
	SocketLabsAPIKey    string `json:"SocketLabsAPIkey"`
	SocketLabsWeight    string `json:"SocketLabsWeight"`
	PostmarkServerToken string `json:"PostmarkServerToken"`
	PostmarkWeight      string `json:"PostmarkWeight"`
	SendgridAPIKey      string `json:"SendgridAPIKey"`
	SendgridWeight      string `json:"SendgridWeight"`
	SparkpostAPIKey     string `json:"SparkpostAPIKey"`
	SparkpostWeight     string `json:"SparkpostWeight"`
}

type StandardizedEvent struct {
	MessageID        string
	Provider         string
	Processed        bool
	ProcessedTime    int64
	Delivered        bool
	DeliveredTime    *int64
	Bounce           bool
	BounceType       string
	BounceTime       *int64
	Deferred         bool
	DeferredCount    int
	LastDeferralTime *int64
	UniqueOpen       bool
	UniqueOpenTime   *int64
	Open             bool
	OpenCount        int
	LastOpenTime     *int64
	Dropped          bool
	DroppedTime      *int64
	DroppedReason    string
}

type ESPCredential struct {
	ProviderName             string
	SendingDomains           []string
	SendgridVerificationKey  string
	SparkpostWebhookUser     string
	SparkpostWebhookPassword string
	SocketlabsSecretKey      string
	PostmarkWebhookUser      string
	PostmarkWebhookPassword  string
	SocketlabsServerID       string
}

func fetchESPCredentials(userID int) (Credentials, error) {
	database.InitDB()
	db := database.GetDB()

	query := `
        SELECT provider_name, 
               socketlabs_server_id, socketlabs_secret_key,
               postmark_webhook_password,
               sendgrid_api_key,
               sparkpost_webhook_password,
			   weight
        FROM email_service_providers
        WHERE user_id = $1
    `
	rows, err := db.Query(query, userID)
	if err != nil {
		return Credentials{}, fmt.Errorf("failed to query ESP credentials: %v", err)
	}
	defer rows.Close()

	creds := Credentials{}
	rowCount := 0

	for rows.Next() {
		rowCount++
		var (
			providerName, socketlabsServerID, socketlabsSecretKey,
			postmarkWebhookPassword, sendgridAPIKey,
			sparkpostWebhookPassword, senderWeight sql.NullString
		)

		err := rows.Scan(
			&providerName,
			&socketlabsServerID, &socketlabsSecretKey,
			&postmarkWebhookPassword,
			&sendgridAPIKey,
			&sparkpostWebhookPassword,
			&senderWeight,
		)
		if err != nil {
			return Credentials{}, fmt.Errorf("failed to scan ESP credential: %v", err)
		}

		log.Printf("Processing provider: %s", providerName.String)

		switch providerName.String {
		case "socketLabs":
			if socketlabsServerID.Valid && socketlabsSecretKey.Valid {
				creds.SocketLabsServerID = socketlabsServerID.String
				creds.SocketLabsAPIKey = socketlabsSecretKey.String
				creds.SocketLabsWeight = senderWeight.String
				log.Printf("Set SocketLabs credentials")
			}
		case "postmark":
			if postmarkWebhookPassword.Valid {
				creds.PostmarkServerToken = postmarkWebhookPassword.String
				creds.PostmarkWeight = senderWeight.String
				log.Printf("Set Postmark credentials")
			}
		case "sendgrid":
			if sendgridAPIKey.Valid {
				creds.SendgridAPIKey = sendgridAPIKey.String
				creds.SendgridWeight = senderWeight.String
				log.Printf("Set SendGrid credentials")
			}
		case "sparkpost":
			if sparkpostWebhookPassword.Valid {
				creds.SparkpostAPIKey = sparkpostWebhookPassword.String
				creds.SparkpostWeight = senderWeight.String
				log.Printf("Set SparkPost credentials")
			}
		default:
			log.Printf("Unknown provider: %s", providerName.String)
		}
	}

	if err := rows.Err(); err != nil {
		return Credentials{}, fmt.Errorf("error iterating over rows: %v", err)
	}

	log.Printf("Processed %d rows", rowCount)
	log.Printf("Resulting credentials: %+v", creds)

	if rowCount == 0 {
		return Credentials{}, fmt.Errorf("no ESP credentials found for user ID %d", userID)
	}

	return creds, nil
}

type ProviderStats struct {
	Name             string
	TotalEvents      int
	DeliveredEvents  int
	BounceEvents     int
	OpenEvents       int
	DeferredEvents   int
	SpamReportEvents int
}

func getProviderStats(db *sql.DB, userID int, daysBack int) ([]ProviderStats, error) {
	query := `
    SELECT 
        esp.provider_name,
        COUNT(*) as total_events,
        SUM(CASE WHEN e.delivered THEN 1 ELSE 0 END) as delivered_events,
        SUM(CASE WHEN e.bounce THEN 1 ELSE 0 END) as bounce_events,
        SUM(CASE WHEN e.open THEN 1 ELSE 0 END) as open_events,
        SUM(CASE WHEN e.deferred THEN 1 ELSE 0 END) as deferred_events,
        SUM(CASE WHEN e.dropped AND e.dropped_reason LIKE '%spam%' THEN 1 ELSE 0 END) as spam_report_events
    FROM 
        events e
    JOIN 
        message_user_associations mua ON e.message_id = mua.message_id
    JOIN 
        email_service_providers esp ON mua.esp_id = esp.esp_id
    WHERE 
        esp.user_id = $1
        AND e.processed_time >= $2
    GROUP BY 
        esp.provider_name
    `

	startDate := time.Now().AddDate(0, 0, -daysBack)
	rows, err := db.Query(query, userID, startDate.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []ProviderStats
	for rows.Next() {
		var s ProviderStats
		if err := rows.Scan(&s.Name, &s.TotalEvents, &s.DeliveredEvents, &s.BounceEvents, &s.OpenEvents, &s.DeferredEvents, &s.SpamReportEvents); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// calculateWeights determines the distribution of emails across different ESP providers
// based on their performance over the last 30 days. It considers multiple factors including
// open rates, successful deliveries, bounces, and spam reports. The function:
//
// 1. Fetches comprehensive event statistics for each provider from the database.
// 2. Calculates a score for each provider based on a weighted formula:
//   - Open rate (50% weight): Higher open rates increase the score significantly.
//   - Success rate (20% weight): Higher delivery rates increase the score.
//   - Bounce rate (30% weight): Higher bounce rates decrease the score.
//   - Spam report rate (20% weight): Higher spam reports decrease the score.
//
// 3. Normalizes these scores into weights that sum to 1,000, providing fine-grained control.
// 4. Adjusts weights to zero for providers with invalid or missing credentials.
// 5. If no provider has a positive score, it distributes weight equally among valid providers.
//
// The resulting weights determine the proportion of emails to be sent through each provider,
// favoring those with better overall performance while maintaining some traffic to all
// valid providers. This approach allows for dynamic load balancing and optimization of
// email deliverability across multiple ESPs, with a strong emphasis on open rates and
// spam prevention.
func calculateWeights(db *sql.DB, userID int, credentials Credentials) (map[string]int, error) {
	stats, err := getProviderStats(db, userID, 30) // Get stats for last 30 days
	if err != nil {
		return nil, err
	}

	weights := make(map[string]int)
	totalScore := 0.0

	for _, s := range stats {
		if s.TotalEvents > 0 {
			openRate := float64(s.OpenEvents) / float64(s.TotalEvents)
			successRate := float64(s.DeliveredEvents) / float64(s.TotalEvents)
			bounceRate := float64(s.BounceEvents) / float64(s.TotalEvents)
			spamRate := float64(s.SpamReportEvents) / float64(s.TotalEvents)

			// Calculate score with appropriate weightings
			score := (openRate * 0.5) + (successRate * 0.2) - (bounceRate * 0.3) - (spamRate * 0.2)
			if score < 0 {
				score = 0 // Ensure the score doesn't go negative
			}

			totalScore += score
			weights[s.Name] = int(score * 1000) // Scale for granularity
		}
	}

	if totalScore > 0 {
		for provider := range weights {
			normalizedWeight := int(float64(weights[provider]) / totalScore * 1000)
			weights[provider] = normalizedWeight
		}
	} else {
		// Assign equal weight to valid providers if totalScore is 0
		validProviders := 0
		for provider := range weights {
			if isValidProvider(provider, credentials) {
				validProviders++
			}
		}

		if validProviders > 0 {
			equalWeight := 1000 / validProviders
			for provider := range weights {
				if isValidProvider(provider, credentials) {
					weights[provider] = equalWeight
				} else {
					weights[provider] = 0
				}
			}
		}
	}

	return weights, nil
}
