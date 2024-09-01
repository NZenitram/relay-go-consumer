package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"relay-go-consumer/database"
)

// func ProcessWebhookMessages(msg *sarama.ConsumerMessage) {
// 	// Initialize the database connection
// 	database.InitDB()
// 	defer database.CloseDB()

// 	// Get the database connection
// 	// db := database.GetDB()

// 	// Use db for your database operations
// 	// ...

// 	log.Println("Application started")
// 	// Rest of your application logic
// }

type UniversalEventSchema struct {
	ID                   int64     `json:"id"`
	EventType            string    `json:"event_type"`
	Provider             string    `json:"provider"`
	TotalCount           int64     `json:"total_count"`
	DailyCount           int64     `json:"daily_count"`
	HourlyCount          int64     `json:"hourly_count"`
	UniqueRecipientCount int64     `json:"unique_recipient_count"`
	UniqueSenderCount    int64     `json:"unique_sender_count"`
	CampaignCount        int64     `json:"campaign_count"`
	DomainCount          int64     `json:"domain_count"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	MessageID            string    `json:"message_id"`
	RecipientEmail       string    `json:"recipient_email"`
	SenderEmail          string    `json:"sender_email"`
	Subject              string    `json:"subject"`
	Timestamp            int64     `json:"timestamp"`
	CampaignID           string    `json:"campaign_id"`
	RecipientDomain      string    `json:"recipient_domain"`
}

// Placeholder function to save events to a database
func saveToDatabase(event interface{}) {
	// Convert event to UniversalEventSchema
	var universalEvent UniversalEventSchema
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	err = json.Unmarshal(eventJSON, &universalEvent)
	if err != nil {
		log.Printf("Error unmarshaling event to UniversalEventSchema: %v", err)
		return
	}

	// Get database connection
	database.InitDB()
	db := database.GetDB()

	// Prepare SQL statement
	stmt, err := db.Prepare(`
        INSERT INTO events (
            event_type, provider, total_count, daily_count, hourly_count,
            unique_recipient_count, unique_sender_count, campaign_count,
            domain_count, created_at, updated_at, message_id, recipient_email,
            sender_email, subject, timestamp, campaign_id, recipient_domain
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
        ) RETURNING id
    `)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return
	}
	defer stmt.Close()

	// Execute the statement
	var id int64
	err = stmt.QueryRow(
		universalEvent.EventType, universalEvent.Provider, universalEvent.TotalCount,
		universalEvent.DailyCount, universalEvent.HourlyCount, universalEvent.UniqueRecipientCount,
		universalEvent.UniqueSenderCount, universalEvent.CampaignCount, universalEvent.DomainCount,
		universalEvent.CreatedAt, universalEvent.UpdatedAt, universalEvent.MessageID,
		universalEvent.RecipientEmail, universalEvent.SenderEmail, universalEvent.Subject,
		universalEvent.Timestamp, universalEvent.CampaignID, universalEvent.RecipientDomain,
	).Scan(&id)

	if err != nil {
		log.Printf("Error inserting event into database: %v", err)
		return
	}

	fmt.Printf("Event saved to database with ID: %d\n", id)

	// TODO: Implement database saving logic
	fmt.Printf("Saving event to database: %+v\n", event)
}
