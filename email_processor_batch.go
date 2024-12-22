package main

import (
	"database/sql"
)

func fetchBatchData(db *sql.DB, batchID int) (BatchInfo, error) {
	var batch BatchInfo

	err := db.QueryRow(`
        SELECT batch_id, total_messages, batch_size, interval_seconds, start_time, end_time,
               current_batch, created_at, updated_at, total_batches, user_id, batches_to_kafka, status
        FROM email_batches
        WHERE batch_id = ?
    `, batchID).Scan(
		&batch.BatchID,
		&batch.TotalMessages,
		&batch.BatchSize,
		&batch.IntervalSeconds,
		&batch.StartTime,
		&batch.EndTime,
		&batch.CurrentBatch,
		&batch.CreatedAt,
		&batch.UpdatedAt,
		&batch.TotalBatches,
		&batch.UserID,
		&batch.BatchesToKafka,
		&batch.Status,
	)

	if err != nil {
		return BatchInfo{}, err
	}

	return batch, nil
}
