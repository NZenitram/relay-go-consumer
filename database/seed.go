package database

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

const (
	numMessages        = 10000000
	sixMonthsInSeconds = 15778800 // Approximately 6 months in seconds (365.25/2 * 24 * 60 * 60)
	batchSize          = 1000
	numWorkers         = 4
)

type ESP struct {
	ESPID     int
	UserID    int
	Provider  string
	Weight    int
	Domain    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Event struct {
	MessageID        string
	Processed        bool
	ProcessedTime    int64
	Delivered        bool
	DeliveredTime    *int64
	Bounce           bool
	BounceType       *string
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
	DroppedReason    *string
	Provider         string
	Metadata         []byte
}

func SeedDB(db *sql.DB) error {
	esps, err := getExistingESPs(db)
	if err != nil {
		return fmt.Errorf("error getting existing ESPs: %v", err)
	}

	totalWeight := 0
	for _, esp := range esps {
		totalWeight += esp.Weight
	}

	endTime := time.Now().UTC()
	startTime := endTime.Add(-time.Duration(sixMonthsInSeconds) * time.Second)

	// Create a channel to distribute work
	jobs := make(chan int, numMessages)
	results := make(chan error, numMessages)

	// Start worker goroutines
	for w := 0; w < numWorkers; w++ {
		go worker(jobs, results, esps, totalWeight, startTime, db)
	}

	// Send jobs to the channel
	for i := 0; i < numMessages; i++ {
		jobs <- i
	}
	close(jobs)

	// Collect results
	for i := 0; i < numMessages; i++ {
		if err := <-results; err != nil {
			return fmt.Errorf("error in worker: %v", err)
		}
	}

	return nil
}

func worker(jobs <-chan int, results chan<- error, esps []ESP, totalWeight int, startTime time.Time, db *sql.DB) {
	for i := range jobs {
		err := func() error {
			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("error starting transaction: %v", err)
			}
			defer tx.Rollback()

			esp := selectESP(esps, totalWeight)
			messageID := fmt.Sprintf("msg_%d_%s_%d", esp.UserID, esp.Provider, i)
			randomTime := startTime.Add(time.Duration(rand.Int63n(sixMonthsInSeconds)) * time.Second)

			// Insert message_user_association
			_, err = tx.Exec(`
                INSERT INTO message_user_associations 
                (message_id, user_id, esp_id, provider, created_at)
                VALUES ($1, $2, $3, $4, $5)
            `, messageID, esp.UserID, esp.ESPID, esp.Provider, randomTime)
			if err != nil {
				return fmt.Errorf("error inserting message_user_association: %v", err)
			}

			// Generate and insert event
			event := generateEvent(messageID, esp.Provider, randomTime.Unix())
			_, err = tx.Exec(`
                INSERT INTO events 
                (message_id, processed, processed_time, delivered, delivered_time, 
                bounce, bounce_type, bounce_time, deferred, deferred_count, 
                last_deferral_time, unique_open, unique_open_time, open, open_count, 
                last_open_time, dropped, dropped_time, dropped_reason, provider, metadata)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
            `, event.MessageID, event.Processed, event.ProcessedTime, event.Delivered, event.DeliveredTime,
				event.Bounce, event.BounceType, event.BounceTime, event.Deferred, event.DeferredCount,
				event.LastDeferralTime, event.UniqueOpen, event.UniqueOpenTime, event.Open, event.OpenCount,
				event.LastOpenTime, event.Dropped, event.DroppedTime, event.DroppedReason, event.Provider, event.Metadata)
			if err != nil {
				return fmt.Errorf("error inserting event: %v", err)
			}

			return tx.Commit()
		}()

		results <- err
	}
}

func getExistingESPs(db *sql.DB) ([]ESP, error) {
	rows, err := db.Query(`
		SELECT esp_id, user_id, provider_name, 
		CASE 
			WHEN provider_name = 'sendgrid' THEN 5
			WHEN provider_name = 'postmark' THEN 3
			WHEN provider_name = 'socketlabs' THEN 2
			WHEN provider_name = 'sparkpost' THEN 4
			ELSE 1
		END as weight,
		sending_domains[1], created_at, updated_at 
		FROM email_service_providers
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var esps []ESP
	for rows.Next() {
		var esp ESP
		err := rows.Scan(&esp.ESPID, &esp.UserID, &esp.Provider, &esp.Weight, &esp.Domain, &esp.CreatedAt, &esp.UpdatedAt)
		if err != nil {
			return nil, err
		}
		esps = append(esps, esp)
	}

	return esps, nil
}

func selectESP(esps []ESP, totalWeight int) ESP {
	r := rand.Intn(totalWeight)
	for _, esp := range esps {
		r -= esp.Weight
		if r < 0 {
			return esp
		}
	}
	return esps[len(esps)-1]
}

func generateEvent(messageID, provider string, processedTime int64) Event {
	event := Event{
		MessageID:     messageID,
		Processed:     true,
		ProcessedTime: processedTime,
		Provider:      provider,
		Metadata:      []byte("{}"),
	}

	// Define provider-specific rates
	var deliveryRate, bounceRate, deferRate, dropRate float64
	switch provider {
	case "sendgrid":
		deliveryRate = 0.92
		bounceRate = 0.05
		deferRate = 0.02
		dropRate = 0.01
	case "postmark":
		deliveryRate = 0.65
		bounceRate = 0.06
		deferRate = 0.03
		dropRate = 0.01
	case "socketlabs":
		deliveryRate = 0.89
		bounceRate = 0.07
		deferRate = 0.03
		dropRate = 0.02
	case "sparkpost":
		deliveryRate = 0.45
		bounceRate = 0.06
		deferRate = 0.03
		dropRate = 0.02
	default:
		deliveryRate = 0.85
		bounceRate = 0.08
		deferRate = 0.04
		dropRate = 0.03
	}

	r := rand.Float64()
	switch {
	case r < deliveryRate:
		event.Delivered = true
		deliveredTime := processedTime + int64(rand.Intn(300)) // Deliver within 5 minutes
		event.DeliveredTime = &deliveredTime

		if rand.Float64() < 0.25 { // 25% of delivered are opened
			event.Open = true
			event.OpenCount = rand.Intn(5) + 1
			lastOpenTime := deliveredTime + int64(rand.Intn(604800)) // Open within a week
			event.LastOpenTime = &lastOpenTime

			if rand.Float64() < 0.9 { // 90% of opens are unique
				event.UniqueOpen = true
				event.UniqueOpenTime = &lastOpenTime
			}
		}

	case r < deliveryRate+bounceRate:
		event.Bounce = true
		bounceTime := processedTime + int64(rand.Intn(300)) // Bounce within 5 minutes
		event.BounceTime = &bounceTime
		bounceType := []string{"hard", "soft", "block"}[rand.Intn(3)]
		event.BounceType = &bounceType

	case r < deliveryRate+bounceRate+deferRate:
		event.Deferred = true
		event.DeferredCount = rand.Intn(5) + 1
		lastDeferralTime := processedTime + int64(rand.Intn(3600)) // Defer within an hour
		event.LastDeferralTime = &lastDeferralTime

	case r < deliveryRate+bounceRate+deferRate+dropRate:
		event.Dropped = true
		droppedTime := processedTime + int64(rand.Intn(300)) // Drop within 5 minutes
		event.DroppedTime = &droppedTime
		droppedReason := []string{"invalid_email", "spam_report", "unsubscribe"}[rand.Intn(3)]
		event.DroppedReason = &droppedReason
	}

	return event
}
