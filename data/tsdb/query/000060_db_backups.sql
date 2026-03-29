-- name: ListDBBackupsByEnv :many
SELECT
  backup_id,
  env,
  backup_type,
  db_name,
  schema_version,
  bucket,
  object_key,
  size_bytes,
  checksum_sha256,
  created_at,
  started_at,
  finished_at,
  status,
  meta
FROM db_backups
WHERE env = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetDBBackupByID :one
SELECT
  backup_id,
  env,
  backup_type,
  db_name,
  schema_version,
  bucket,
  object_key,
  size_bytes,
  checksum_sha256,
  created_at,
  started_at,
  finished_at,
  status,
  meta
FROM db_backups
WHERE backup_id = $1;
