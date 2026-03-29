package repository

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("dbbackups repository: not found")
	ErrNotConfigured = errors.New("dbbackups repository: not configured")
)

type DBBackup struct {
	BackupID       string
	Env            string
	BackupType     string
	DBName         string
	SchemaVersion  string
	Bucket         string
	ObjectKey      string
	SizeBytes      int64
	ChecksumSHA256 string
	CreatedAt      time.Time
	StartedAt      *time.Time
	FinishedAt     *time.Time
	Status         string
	Meta           map[string]any
}

type Reader interface {
	ListByEnv(ctx context.Context, env string, limit, offset int) ([]DBBackup, error)
	GetByID(ctx context.Context, backupID string) (DBBackup, error)
}
