-- name: InsertArtifact :exec
INSERT INTO artifacts (
  artifact_id,
  workflow_run_id,
  task_id,
  attempt_no,
  bucket,
  object_key,
  kind,
  filename,
  content_type,
  size_bytes,
  checksum_sha256,
  etag,
  status,
  tags
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
);

-- name: ListArtifactsByWorkflow :many
SELECT
  artifact_id,
  workflow_run_id,
  task_id,
  attempt_no,
  kind,
  filename,
  content_type,
  size_bytes,
  status,
  created_at,
  tags
FROM artifacts
WHERE workflow_run_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetArtifactByID :one
SELECT *
FROM artifacts
WHERE artifact_id = $1;
