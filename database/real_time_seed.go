package database

import (
	"database/sql"
	"fmt"
	"time"
)

func RealTimeSeedDB(db *sql.DB) error {
	fmt.Println("Starting real-time database seeding...")

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

	// Get the last event time and perform backfill if necessary
	lastEventTime, err := getLastEventTime(db)
	if err != nil {
		return fmt.Errorf("error getting last event time: %v", err)
	}

	now := time.Now().UTC()
	if lastEventTime.Before(now.Add(-24 * time.Hour)) {
		err = backfillEvents(db, lastEventTime, now, esps, totalWeight)
		if err != nil {
			return fmt.Errorf("error during backfill: %v", err)
		}
	}

	// Create channels for job distribution and result collection
	jobs := make(chan int, batchSize*numWorkers)
	results := make(chan error, batchSize*numWorkers)

	// Start worker goroutines
	for w := 0; w < numWorkers; w++ {
		go worker(w, jobs, results, esps, totalWeight, now, db)
	}

	// Create a ticker for controlling the seeding rate
	ticker := time.NewTicker(time.Millisecond * 10) // Adjust this value to control the seeding rate
	defer ticker.Stop()

	// Counter for generated messages
	messageCounter := 0

	// Start time for rate calculation
	startTime := time.Now()

	for range ticker.C {
		// Send a new job
		jobs <- messageCounter
		messageCounter++

		// Process the result
		select {
		case err := <-results:
			if err != nil {
				fmt.Printf("Error in worker: %v\n", err)
			}
		default:
			// No result available, continue
		}

		// Print statistics every 10000 messages
		if messageCounter%10000 == 0 {
			elapsed := time.Since(startTime)
			rate := float64(messageCounter) / elapsed.Seconds()
			fmt.Printf("Generated %d messages. Current rate: %.2f messages/second\n", messageCounter, rate)
		}
	}

	return nil
}

func getLastEventTime(db *sql.DB) (time.Time, error) {
	var lastTime time.Time
	err := db.QueryRow("SELECT MAX(processed_time) FROM events").Scan(&lastTime)
	if err != nil && err != sql.ErrNoRows {
		return time.Time{}, err
	}
	if lastTime.IsZero() {
		// If no events exist, return a time 24 hours ago
		return time.Now().UTC().Add(-24 * time.Hour), nil
	}
	return lastTime, nil
}

func backfillEvents(db *sql.DB, startTime, endTime time.Time, esps []ESP, totalWeight int) error {
	fmt.Println("Starting backfill...")

	jobs := make(chan int, batchSize*numWorkers)
	results := make(chan error, batchSize*numWorkers)

	// Start worker goroutines
	for w := 0; w < numWorkers; w++ {
		go worker(w, jobs, results, esps, totalWeight, startTime, db)
	}

	// Calculate the number of events to generate
	duration := endTime.Sub(startTime)
	numEvents := int(duration.Seconds()) // Assuming 1 event per second, adjust as needed

	// Send jobs to the channel
	go func() {
		for i := 0; i < numEvents; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	// Collect results
	for i := 0; i < numEvents; i++ {
		if err := <-results; err != nil {
			fmt.Printf("Error in worker during backfill: %v\n", err)
		}
		if i > 0 && i%10000 == 0 {
			fmt.Printf("Backfilled %d/%d events\n", i, numEvents)
		}
	}

	fmt.Println("Backfill complete.")
	return nil
}
