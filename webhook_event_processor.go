package main

import (
	"fmt"
	"relay-go-consumer/database"
)

func saveStandardizedEvent(event StandardizedEvent) error {
	database.InitDB()
	db := database.GetDB()

	stmt, err := db.Prepare(`
        INSERT INTO events (
            message_id, provider, processed, processed_time, delivered, delivered_time,
            bounce, bounce_type, bounce_time, deferred, deferred_count,
            last_deferral_time, unique_open, unique_open_time, open, open_count, last_open_time,
            dropped, dropped_time, dropped_reason
        ) VALUES (
            ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
        ) ON DUPLICATE KEY UPDATE
            provider = VALUES(provider),
            processed = VALUES(processed),
            processed_time = VALUES(processed_time),
            delivered = VALUES(delivered),
            delivered_time = COALESCE(VALUES(delivered_time), delivered_time),
            bounce = VALUES(bounce),
            bounce_type = COALESCE(VALUES(bounce_type), bounce_type),
            bounce_time = COALESCE(VALUES(bounce_time), bounce_time),
            deferred = VALUES(deferred),
            deferred_count = deferred_count + VALUES(deferred_count),
            last_deferral_time = COALESCE(VALUES(last_deferral_time), last_deferral_time),
            unique_open = VALUES(unique_open),
            unique_open_time = COALESCE(VALUES(unique_open_time), unique_open_time),
            open = VALUES(open),
            open_count = open_count + VALUES(open_count),
            last_open_time = COALESCE(VALUES(last_open_time), last_open_time),
            dropped = VALUES(dropped),
            dropped_time = COALESCE(VALUES(dropped_time), dropped_time),
            dropped_reason = COALESCE(VALUES(dropped_reason), dropped_reason)
    `)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		event.MessageID,
		event.Provider,
		event.Processed,
		event.ProcessedTime,
		event.Delivered,
		event.DeliveredTime,
		event.Bounce,
		event.BounceType,
		event.BounceTime,
		event.Deferred,
		event.DeferredCount,
		event.LastDeferralTime,
		event.UniqueOpen,
		event.UniqueOpenTime,
		event.Open,
		event.OpenCount,
		event.LastOpenTime,
		event.Dropped,
		event.DroppedTime,
		event.DroppedReason,
	)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %v", err)
	}

	return nil
}
