-- name: InsertOutboxEvent :exec
INSERT INTO outbox_events (
  event_type,
  partition_key,
  payload_json
) VALUES (
  $1, $2, $3
);

-- name: ListPendingOutboxEvents :many
SELECT id, event_type, partition_key, payload_json, created_at, attempts
FROM outbox_events
WHERE published_at IS NULL
  AND (next_publish_at IS NULL OR next_publish_at <= now())
ORDER BY created_at
LIMIT $1;

-- name: MarkOutboxEventPublished :exec
UPDATE outbox_events
SET published_at = $2
WHERE id = $1;

-- name: MarkOutboxEventFailed :exec
UPDATE outbox_events
SET attempts = attempts + 1,
    last_error = $2,
    next_publish_at = $3
WHERE id = $1;
