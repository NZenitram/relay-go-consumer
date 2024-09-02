package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/lib/pq"
)

const (
	eventsPerDay   = 1000
	daysToGenerate = 30
)

var userIDs = []int{1, 2, 3, 4, 5} // Assuming these are the user IDs from your users table

// EventProbability represents the probability and count of each event type
type EventProbability struct {
	eventType   string
	probability float64
	count       int
}

// Define event probabilities based on typical email deliverability statistics
var eventProbabilities = []EventProbability{
	{"delivered", 0.857, 0},   // 85.7% average delivery rate
	{"opened", 0.214, 0},      // 21.4% open rate (25% of delivered)
	{"clicked", 0.0321, 0},    // 3.21% click rate (15% of opened)
	{"bounced", 0.023, 0},     // 2.3% bounce rate
	{"spam", 0.063, 0},        // 6.3% spam rate
	{"unsubscribe", 0.002, 0}, // 0.2% unsubscribe rate
	{"deferred", 0.05, 0},     // 5% deferred rate
	{"dropped", 0.03, 0},      // 3% dropped rate
}

func calculateEventCounts(totalEvents int) {
	for i := range eventProbabilities {
		eventProbabilities[i].count = int(float64(totalEvents) * eventProbabilities[i].probability)
		if eventProbabilities[i].count == 0 {
			eventProbabilities[i].count = 1 // Ensure at least one event of each type
		}
	}
}

func SeedDB() error {
	InitDB()
	db := GetDB()
	defer db.Close()

	err := seedEvents(db)
	if err != nil {
		return fmt.Errorf("failed to seed events: %v", err)
	}

	fmt.Println("Event seeding completed successfully.")
	printEventCounts()
	return nil
}

func seedEvents(db *sql.DB) error {
	startDate := time.Now().AddDate(0, 0, -daysToGenerate)
	totalEvents := eventsPerDay * daysToGenerate * 4 // 4 providers
	calculateEventCounts(totalEvents)

	for _, ep := range eventProbabilities {
		for i := 0; i < ep.count; i++ {
			day := rand.Intn(daysToGenerate)
			currentDate := startDate.AddDate(0, 0, day)

			if err := seedPostmarkEvents(db, currentDate, ep.eventType); err != nil {
				return err
			}
			if err := seedSendgridEvents(db, currentDate, ep.eventType); err != nil {
				return err
			}
			if err := seedSocketlabsEvents(db, currentDate, ep.eventType); err != nil {
				return err
			}
			if err := seedSparkpostEvents(db, currentDate, ep.eventType); err != nil {
				return err
			}
		}
	}

	return nil
}

// func getRandomEvent() string {
// 	r := rand.Float64()
// 	cumulativeProbability := 0.0
// 	for i, ep := range eventProbabilities {
// 		cumulativeProbability += ep.probability
// 		if r <= cumulativeProbability {
// 			eventProbabilities[i].count++
// 			return ep.eventType
// 		}
// 	}
// 	return eventProbabilities[len(eventProbabilities)-1].eventType
// }

func seedPostmarkEvents(db *sql.DB, date time.Time, eventType string) error {
	stmt, err := db.Prepare(`
        INSERT INTO public.postmark_events (
            record_type, server_id, message_id, recipient, tag, delivered_at, details, metadata,
            provider, event_type, event_data, accept_encoding, content_length, content_type, expect,
            user_agent, x_forwarded_for, x_forwarded_host, x_forwarded_proto, x_pm_retries_remaining,
            x_pm_webhook_event_id, x_pm_webhook_trace_id, auth_header, "timestamp", user_id
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
            $21, $22, $23, $24, $25
        )
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	eventTime := date.Add(time.Duration(rand.Intn(24*60*60)) * time.Second)
	eventData, _ := json.Marshal(map[string]string{"key": "value"})
	_, err = stmt.Exec(
		"Inbound", rand.Intn(1000), fmt.Sprintf("msg_%d", rand.Int()),
		fmt.Sprintf("recipient%d@example.com", rand.Intn(1000)), "tag",
		eventTime, "details", eventData, "Postmark", eventType,
		eventData, pq.Array([]string{"gzip"}), pq.Array([]string{"100"}),
		pq.Array([]string{"application/json"}), pq.Array([]string{}),
		pq.Array([]string{"User-Agent"}), pq.Array([]string{"127.0.0.1"}),
		pq.Array([]string{"example.com"}), pq.Array([]string{"https"}),
		pq.Array([]string{"3"}), pq.Array([]string{"event-id"}),
		pq.Array([]string{"trace-id"}), "auth-header", eventTime.Unix(),
		userIDs[rand.Intn(len(userIDs))],
	)
	if err != nil {
		return err
	}
	return nil
}

func seedSendgridEvents(db *sql.DB, date time.Time, eventType string) error {
	stmt, err := db.Prepare(`
        INSERT INTO public.sendgrid_events (
            provider, email, "timestamp", smtp_id, event_type, category, sg_event_id, sg_message_id,
            accept_encoding, content_length, content_type, user_agent, x_forwarded_for,
            x_forwarded_host, x_forwarded_proto, x_twilio_email_event_webhook_signature,
            x_twilio_email_event_webhook_timestamp, user_id
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
        )
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	eventTime := date.Add(time.Duration(rand.Intn(24*60*60)) * time.Second)
	_, err = stmt.Exec(
		"Sendgrid", fmt.Sprintf("recipient%d@example.com", rand.Intn(1000)),
		eventTime.Unix(), fmt.Sprintf("smtp_%d", rand.Int()),
		eventType, pq.Array([]string{"category1", "category2"}),
		fmt.Sprintf("sg_event_%d", rand.Int()), fmt.Sprintf("sg_message_%d", rand.Int()),
		pq.Array([]string{"gzip"}), pq.Array([]string{"100"}),
		pq.Array([]string{"application/json"}), pq.Array([]string{"User-Agent"}),
		pq.Array([]string{"127.0.0.1"}), pq.Array([]string{"example.com"}),
		pq.Array([]string{"https"}), pq.Array([]string{"signature"}),
		pq.Array([]string{eventTime.Format(time.RFC3339)}),
		userIDs[rand.Intn(len(userIDs))],
	)
	if err != nil {
		return err
	}
	return nil
}

func seedSocketlabsEvents(db *sql.DB, date time.Time, eventType string) error {
	stmt, err := db.Prepare(`
        INSERT INTO public.socketlabs_events (
            event_type, date_time, mailing_id, message_id, address, server_id, subaccount_id,
            ip_pool_id, secret_key, event_data, accept_encoding, content_length, content_type,
            user_agent, x_forwarded_for, x_forwarded_host, x_forwarded_proto,
            x_socketlabs_signature, "timestamp", user_id
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
        )
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	eventTime := date.Add(time.Duration(rand.Intn(24*60*60)) * time.Second)
	eventData, _ := json.Marshal(map[string]string{"key": "value"})
	_, err = stmt.Exec(
		eventType,
		eventTime,
		fmt.Sprintf("mailing_%d", rand.Int()), fmt.Sprintf("msg_%d", rand.Int()),
		fmt.Sprintf("recipient%d@example.com", rand.Intn(1000)),
		rand.Intn(1000), rand.Intn(100), rand.Intn(10),
		"secret-key", eventData,
		pq.Array([]string{"gzip"}), pq.Array([]string{"100"}),
		pq.Array([]string{"application/json"}), pq.Array([]string{"User-Agent"}),
		pq.Array([]string{"127.0.0.1"}), pq.Array([]string{"example.com"}),
		pq.Array([]string{"https"}), pq.Array([]string{"signature"}),
		eventTime.Unix(), userIDs[rand.Intn(len(userIDs))],
	)
	if err != nil {
		return fmt.Errorf("error inserting event: %v", err)
	}
	return nil
}

func seedSparkpostEvents(db *sql.DB, date time.Time, eventType string) error {
	stmt, err := db.Prepare(`
        INSERT INTO public.sparkpost_events (
            event_type, message_id, transmission_id, event_data, accept_encoding, content_length,
            content_type, user_agent, x_forwarded_for, x_forwarded_host, x_forwarded_proto,
            x_sparkpost_signature, "timestamp", rcpt_to, ip_address, event_id, auth_header, user_id
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
        )
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	eventTime := date.Add(time.Duration(rand.Intn(24*60*60)) * time.Second)
	eventData, _ := json.Marshal(map[string]string{"key": "value"})
	_, err = stmt.Exec(
		eventType,
		fmt.Sprintf("msg_%d", rand.Int()),
		fmt.Sprintf("trans_%d", rand.Int()),
		eventData,
		pq.Array([]string{"gzip"}),
		pq.Array([]string{"100"}),
		pq.Array([]string{"application/json"}),
		pq.Array([]string{"User-Agent"}),
		pq.Array([]string{"127.0.0.1"}),
		pq.Array([]string{"example.com"}),
		pq.Array([]string{"https"}),
		pq.Array([]string{"signature"}),
		eventTime.Unix(),
		fmt.Sprintf("recipient%d@example.com", rand.Intn(1000)),
		fmt.Sprintf("192.168.0.%d", rand.Intn(256)),
		fmt.Sprintf("event_%d", rand.Int()),
		"auth-header",
		userIDs[rand.Intn(len(userIDs))],
	)
	if err != nil {
		return err
	}
	return nil
}

func printEventCounts() {
	fmt.Println("Event counts:")
	for _, ep := range eventProbabilities {
		fmt.Printf("%s: %d\n", ep.eventType, ep.count)
	}
}
