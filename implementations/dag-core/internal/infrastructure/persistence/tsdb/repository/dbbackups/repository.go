package dbbackups

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	apprepository "dag-observatory/dag-core/internal/application/dbbackups/repository"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Repository struct {
	queries query.Querier
}

func NewRepository(client *tsdb.Client) *Repository {
	var q query.Querier
	if client != nil && client.Pool != nil {
		q = query.New(client.Pool)
	}
	return NewRepositoryWithQuerier(q)
}

func NewRepositoryWithQuerier(queries query.Querier) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) ListByEnv(ctx context.Context, env string, limit, offset int) ([]apprepository.DBBackup, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListDBBackupsByEnv(ctx, query.ListDBBackupsByEnvParams{
		Env:    env,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, mapError(err)
	}

	out := make([]apprepository.DBBackup, 0, len(rows))
	for _, row := range rows {
		mapped, err := fromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}
	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, backupID string) (apprepository.DBBackup, error) {
	if err := r.validate(); err != nil {
		return apprepository.DBBackup{}, err
	}
	parsedID, err := uuid.Parse(backupID)
	if err != nil {
		return apprepository.DBBackup{}, fmt.Errorf("dbbackups repository: invalid backup_id: %w", err)
	}
	row, err := r.queries.GetDBBackupByID(ctx, toPgUUID(parsedID))
	if err != nil {
		return apprepository.DBBackup{}, mapError(err)
	}
	return fromRow(row)
}

func (r *Repository) validate() error {
	if r == nil || r.queries == nil {
		return apprepository.ErrNotConfigured
	}
	return nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %v", apprepository.ErrNotFound, err)
	}
	return err
}

func fromRow(row query.DbBackup) (apprepository.DBBackup, error) {
	id, err := fromPgUUID(row.BackupID)
	if err != nil {
		return apprepository.DBBackup{}, err
	}
	createdAt, err := fromPgTime(row.CreatedAt)
	if err != nil {
		return apprepository.DBBackup{}, err
	}
	meta, err := parseJSONMap(row.Meta)
	if err != nil {
		return apprepository.DBBackup{}, err
	}

	startedAt, err := fromNullablePgTime(row.StartedAt)
	if err != nil {
		return apprepository.DBBackup{}, err
	}
	finishedAt, err := fromNullablePgTime(row.FinishedAt)
	if err != nil {
		return apprepository.DBBackup{}, err
	}

	return apprepository.DBBackup{
		BackupID:       id.String(),
		Env:            row.Env,
		BackupType:     row.BackupType,
		DBName:         row.DbName,
		SchemaVersion:  row.SchemaVersion.String,
		Bucket:         row.Bucket,
		ObjectKey:      row.ObjectKey,
		SizeBytes:      row.SizeBytes,
		ChecksumSHA256: row.ChecksumSha256.String,
		CreatedAt:      createdAt,
		StartedAt:      startedAt,
		FinishedAt:     finishedAt,
		Status:         row.Status,
		Meta:           meta,
	}, nil
}

func parseJSONMap(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if out == nil {
		return map[string]any{}, nil
	}
	return out, nil
}

func fromPgUUID(v pgtype.UUID) (uuid.UUID, error) {
	if !v.Valid {
		return uuid.UUID{}, fmt.Errorf("dbbackups repository: uuid is NULL")
	}
	return uuid.UUID(v.Bytes), nil
}

func toPgUUID(v uuid.UUID) pgtype.UUID {
	var bytes [16]byte
	copy(bytes[:], v[:])
	return pgtype.UUID{
		Bytes: bytes,
		Valid: true,
	}
}

func fromPgTime(v pgtype.Timestamptz) (time.Time, error) {
	if !v.Valid {
		return time.Time{}, fmt.Errorf("dbbackups repository: timestamp is NULL")
	}
	return v.Time, nil
}

func fromNullablePgTime(v pgtype.Timestamptz) (*time.Time, error) {
	if !v.Valid {
		return nil, nil
	}
	tm := v.Time
	return &tm, nil
}

var _ apprepository.Reader = (*Repository)(nil)
