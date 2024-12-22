package main

import (
	"database/sql"
	"time"
)

type ProviderStats struct {
	Name             string
	TotalEvents      int
	DeliveredEvents  int
	BounceEvents     int
	OpenEvents       int
	DeferredEvents   int
	SpamReportEvents int
}

func getProviderStats(db *sql.DB, userID int, startTime, endTime time.Time) ([]ProviderStats, error) {
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
        esp.user_id = ?
        AND e.processed_time >= ?
        AND e.processed_time < ?
    GROUP BY 
        esp.provider_name
    `

	rows, err := db.Query(query, userID, startTime.Unix(), endTime.Unix())
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
func calculateWeightsForTimeRange(db *sql.DB, userID int, credentials Credentials, startTime, endTime time.Time) (map[string]int, error) {
	stats, err := getProviderStats(db, userID, startTime, endTime)
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
