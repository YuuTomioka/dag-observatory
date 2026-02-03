package tsdb

import (
	"context"
	"fmt"
	"time"
)

type DBBackup struct {
	BackupID      string
	Env           string
	BackupType    string
	DBName        string
	SchemaVersion string
	Bucket        string
	ObjectKey     string
	SizeBytes     int64
	ChecksumSHA256 string
	CreatedAt     time.Time
	StartedAt     *time.Time
	FinishedAt    *time.Time
	Status        string
	Meta          map[string]any
}

type DBBackupsRepository struct {
	pool *Client
}

func NewDBBackupsRepository(client *Client) *DBBackupsRepository {
	return &DBBackupsRepository{pool: client}
}

func (r *DBBackupsRepository) ListByEnv(ctx context.Context, env string, limit, offset int) ([]DBBackup, error) {
	if r == nil || r.pool == nil || r.pool.Pool == nil {
		return nil, fmt.Errorf("tsdb: client not configured")
	}
	rows, err := r.pool.Pool.Query(ctx, `
SELECT backup_id, env, backup_type, db_name, COALESCE(schema_version,''), bucket, object_key,
       size_bytes, COALESCE(checksum_sha256,''), created_at, started_at, finished_at, status, meta
FROM db_backups
WHERE env=$1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`, env, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]DBBackup, 0, limit)
	for rows.Next() {
		var b DBBackup
		if err := rows.Scan(
			&b.BackupID,
			&b.Env,
			&b.BackupType,
			&b.DBName,
			&b.SchemaVersion,
			&b.Bucket,
			&b.ObjectKey,
			&b.SizeBytes,
			&b.ChecksumSHA256,
			&b.CreatedAt,
			&b.StartedAt,
			&b.FinishedAt,
			&b.Status,
			&b.Meta,
		); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *DBBackupsRepository) GetByID(ctx context.Context, backupID string) (DBBackup, error) {
	if r == nil || r.pool == nil || r.pool.Pool == nil {
		return DBBackup{}, fmt.Errorf("tsdb: client not configured")
	}
	var b DBBackup
	err := r.pool.Pool.QueryRow(ctx, `
SELECT backup_id, env, backup_type, db_name, COALESCE(schema_version,''), bucket, object_key,
       size_bytes, COALESCE(checksum_sha256,''), created_at, started_at, finished_at, status, meta
FROM db_backups
WHERE backup_id=$1
`, backupID).Scan(
		&b.BackupID,
		&b.Env,
		&b.BackupType,
		&b.DBName,
		&b.SchemaVersion,
		&b.Bucket,
		&b.ObjectKey,
		&b.SizeBytes,
		&b.ChecksumSHA256,
		&b.CreatedAt,
		&b.StartedAt,
		&b.FinishedAt,
		&b.Status,
		&b.Meta,
	)
	return b, err
}
