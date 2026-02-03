-- name: InsertProcessedEvent :one
INSERT INTO processed_events (
  event_id,
  task_id,
  attempt_no
) VALUES (
  $1, $2, $3
)
ON CONFLICT DO NOTHING
RETURNING event_id;
