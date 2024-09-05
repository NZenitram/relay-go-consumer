package main

import (
	"database/sql"
	"relay-go-consumer/database"
)

func saveStandardizedEvent(event StandardizedEvent) error {
	database.InitDB()
	db := database.GetDB()

	// First, try to update an existing record
	stmt, err := db.Prepare(`
        UPDATE events SET
            provider = $2,
            processed = $3,
            processed_time = $4,
            delivered = $5,
            delivered_time = COALESCE($6, delivered_time),
            bounce = $7,
            bounce_type = COALESCE($8, bounce_type),
            bounce_time = COALESCE($9, bounce_time),
            deferred = $10,
            deferred_count = deferred_count + $11,
            last_deferral_time = COALESCE($12, last_deferral_time),
            unique_open = $13,
            unique_open_time = COALESCE($14, unique_open_time),
            open = $15,
            open_count = open_count + $16,
            last_open_time = COALESCE($17, last_open_time),
            dropped = $18,
            dropped_time = COALESCE($19, dropped_time),
            dropped_reason = COALESCE($20, dropped_reason)
        WHERE message_id = $1
        RETURNING message_id
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var updatedMessageID string
	err = stmt.QueryRow(
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
	).Scan(&updatedMessageID)

	if err == sql.ErrNoRows {
		// If no existing record was updated, insert a new one
		insertStmt, err := db.Prepare(`
            INSERT INTO events (
                message_id, provider, processed, processed_time, delivered, delivered_time,
                bounce, bounce_type, bounce_time, deferred, deferred_count,
                last_deferral_time, unique_open, unique_open_time, open, open_count, last_open_time,
                dropped, dropped_time, dropped_reason
            ) VALUES (
                $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
            )
        `)
		if err != nil {
			return err
		}
		defer insertStmt.Close()

		_, err = insertStmt.Exec(
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
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}
