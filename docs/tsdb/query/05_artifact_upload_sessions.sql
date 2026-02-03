-- name: InsertUploadSession :exec
INSERT INTO artifact_upload_sessions (
  upload_id,
  workflow_run_id,
  task_id,
  attempt_no,
  bucket,
  object_key,
  kind,
  filename,
  content_type,
  status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);

-- name: MarkUploadSessionUploaded :exec
UPDATE artifact_upload_sessions
SET status = $2,
    uploaded_at = $3
WHERE upload_id = $1;

-- name: MarkUploadSessionFailed :exec
UPDATE artifact_upload_sessions
SET status = $2,
    error = $3
WHERE upload_id = $1;

-- name: ListExpiredUploadSessions :many
SELECT upload_id, bucket, object_key
FROM artifact_upload_sessions
WHERE status = 'initiated'
  AND created_at < $1
LIMIT $2;

-- name: MarkUploadSessionExpired :exec
UPDATE artifact_upload_sessions
SET status = 'expired'
WHERE upload_id = $1;
