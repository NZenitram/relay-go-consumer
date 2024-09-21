package database

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

const (
	numMessages        = 100000000
	sixMonthsInSeconds = 31557600 // Approximately 12 months in seconds (365.25/2 * 24 * 60 * 60)
	batchSize          = 10000
	numWorkers         = 24
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
	fmt.Println("Starting database seeding...")

	esps, err := getExistingESPs(db)
	if err != nil {
		return fmt.Errorf("error getting existing ESPs: %v", err)
	}
	fmt.Printf("Found %d ESPs\n", len(esps))

	if len(esps) == 0 {
		return fmt.Errorf("no ESPs found in the database")
	}

	totalWeight := 0
	for _, esp := range esps {
		totalWeight += esp.Weight
	}

	endTime := time.Now().UTC()
	startTime := endTime.Add(-time.Duration(sixMonthsInSeconds) * time.Second)

	// Create a channel to distribute work
	jobs := make(chan int, batchSize*numWorkers)
	results := make(chan error, batchSize*numWorkers)

	// Start worker goroutines
	for w := 0; w < numWorkers; w++ {
		go worker(w, jobs, results, esps, totalWeight, startTime, db)
	}

	// Send jobs to the channel
	go func() {
		for i := 0; i < numMessages; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	// Collect results
	successCount := 0
	errorCount := 0
	startProcessingTime := time.Now()

	for i := 0; i < numMessages; i++ {
		if err := <-results; err != nil {
			fmt.Printf("Error in worker: %v\n", err)
			errorCount++
		} else {
			successCount++
		}

		if i > 0 && i%10000 == 0 {
			elapsed := time.Since(startProcessingTime)
			rate := float64(i) / elapsed.Seconds()
			fmt.Printf("Processed %d messages (%.2f%%)... Current rate: %.2f messages/second\n",
				i, float64(i)/float64(numMessages)*100, rate)
		}
	}

	totalTime := time.Since(startProcessingTime)
	overallRate := float64(numMessages) / totalTime.Seconds()

	fmt.Printf("Seeding complete.\n")
	fmt.Printf("Total time: %v\n", totalTime)
	fmt.Printf("Successfully inserted: %d\n", successCount)
	fmt.Printf("Errors: %d\n", errorCount)
	fmt.Printf("Overall rate: %.2f messages/second\n", overallRate)

	return nil
}

func worker(id int, jobs <-chan int, results chan<- error, esps []ESP, totalWeight int, startTime time.Time, db *sql.DB) {
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

			// Insert into events table
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

			// Insert into TimescaleDB tables
			if event.Processed {
				_, err = tx.Exec(`
                    INSERT INTO processed_events (time, message_id, user_id, esp_id, provider)
                    VALUES ($1, $2, $3, $4, $5)
                `, time.Unix(event.ProcessedTime, 0), event.MessageID, esp.UserID, esp.ESPID, event.Provider)
				if err != nil {
					return fmt.Errorf("error inserting processed event: %v", err)
				}
			}

			if event.Delivered {
				_, err = tx.Exec(`
                    INSERT INTO delivered_events (time, message_id, user_id, esp_id, provider)
                    VALUES ($1, $2, $3, $4, $5)
                `, time.Unix(*event.DeliveredTime, 0), event.MessageID, esp.UserID, esp.ESPID, event.Provider)
				if err != nil {
					return fmt.Errorf("error inserting delivered event: %v", err)
				}
			}

			if event.Bounce {
				_, err = tx.Exec(`
                    INSERT INTO bounce_events (time, message_id, user_id, esp_id, provider, bounce_type)
                    VALUES ($1, $2, $3, $4, $5, $6)
                `, time.Unix(*event.BounceTime, 0), event.MessageID, esp.UserID, esp.ESPID, event.Provider, *event.BounceType)
				if err != nil {
					return fmt.Errorf("error inserting bounce event: %v", err)
				}
			}

			if event.Deferred {
				_, err = tx.Exec(`
                    INSERT INTO deferred_events (time, message_id, user_id, esp_id, provider, deferred_count)
                    VALUES ($1, $2, $3, $4, $5, $6)
                `, time.Unix(*event.LastDeferralTime, 0), event.MessageID, esp.UserID, esp.ESPID, event.Provider, event.DeferredCount)
				if err != nil {
					return fmt.Errorf("error inserting deferred event: %v", err)
				}
			}

			if event.Open {
				_, err = tx.Exec(`
                    INSERT INTO open_events (time, message_id, user_id, esp_id, provider, open_count)
                    VALUES ($1, $2, $3, $4, $5, $6)
                `, time.Unix(*event.LastOpenTime, 0), event.MessageID, esp.UserID, esp.ESPID, event.Provider, event.OpenCount)
				if err != nil {
					return fmt.Errorf("error inserting open event: %v", err)
				}
			}

			if event.Dropped {
				_, err = tx.Exec(`
                    INSERT INTO dropped_events (time, message_id, user_id, esp_id, provider, dropped_reason)
                    VALUES ($1, $2, $3, $4, $5, $6)
                `, time.Unix(*event.DroppedTime, 0), event.MessageID, esp.UserID, esp.ESPID, event.Provider, *event.DroppedReason)
				if err != nil {
					return fmt.Errorf("error inserting dropped event: %v", err)
				}
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
		WHERE user_id = 5
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
		deliveryRate = 0.89
		bounceRate = 0.06
		deferRate = 0.03
		dropRate = 0.01
	case "socketlabs":
		deliveryRate = 0.91
		bounceRate = 0.07
		deferRate = 0.03
		dropRate = 0.02
	case "sparkpost":
		deliveryRate = 0.94
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
	processedDateTime := time.Unix(processedTime, 0)

	// Adjust rates based on day of week and time of day
	if isWeekend(processedDateTime) {
		deliveryRate *= 0.8
		bounceRate *= 1.2
		deferRate *= 1.1
		dropRate *= 1.1
	}

	hourAdjustment := getHourAdjustment(processedDateTime)
	deliveryRate *= hourAdjustment
	bounceRate *= (2 - hourAdjustment)
	deferRate *= (2 - hourAdjustment)
	dropRate *= (2 - hourAdjustment)

	switch {
	case r < deliveryRate:
		event.Delivered = true
		deliveredTime := processedTime + int64(rand.Intn(300)) // Deliver within 5 minutes
		event.DeliveredTime = &deliveredTime

		if rand.Float64() < getOpenProbability(time.Unix(deliveredTime, 0)) {
			event.Open = true
			event.OpenCount = rand.Intn(5) + 1
			lastOpenTime := getRealisticOpenTime(deliveredTime)
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

func isWeekend(t time.Time) bool {
	day := t.Weekday()
	return day == time.Saturday || day == time.Sunday
}

func getHourAdjustment(t time.Time) float64 {
	hour := t.Hour()
	switch {
	case hour >= 7 && hour < 10:
		return 1.3 // Morning peak
	case hour >= 11 && hour < 14:
		return 1.2 // Midday
	case hour >= 15 && hour < 18:
		return 1.1 // Afternoon
	case hour >= 19 && hour < 23:
		return 0.9 // Evening
	default:
		return 0.7 // Late night/early morning
	}
}

func getOpenProbability(t time.Time) float64 {
	baseProbability := 0.25
	if isWeekend(t) {
		baseProbability *= 0.8
	}
	return baseProbability * getHourAdjustment(t)
}

func getRealisticOpenTime(deliveredTime int64) int64 {
	deliveredDateTime := time.Unix(deliveredTime, 0)

	// 50% chance to open within 1 hour, 30% within 24 hours, 20% within a week
	r := rand.Float64()
	var openDelay int64
	switch {
	case r < 0.5:
		openDelay = rand.Int63n(3600) // Within 1 hour
	case r < 0.8:
		openDelay = rand.Int63n(86400) // Within 24 hours
	default:
		openDelay = rand.Int63n(604800) // Within a week
	}

	openDateTime := deliveredDateTime.Add(time.Duration(openDelay) * time.Second)

	// Adjust open time to more realistic hours
	for openDateTime.Hour() >= 23 || openDateTime.Hour() < 6 {
		openDateTime = openDateTime.Add(time.Hour)
	}

	return openDateTime.Unix()
}
