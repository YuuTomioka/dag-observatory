package dto

import "time"

type DBBackupItem struct {
	BackupID      string         `json:"backup_id"`
	Env           string         `json:"env"`
	BackupType    string         `json:"backup_type"`
	DBName        string         `json:"db_name"`
	SchemaVersion string         `json:"schema_version,omitempty"`
	SizeBytes     int64          `json:"size_bytes"`
	ChecksumSHA256 string        `json:"checksum_sha256,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	StartedAt     *time.Time     `json:"started_at,omitempty"`
	FinishedAt    *time.Time     `json:"finished_at,omitempty"`
	Status        string         `json:"status"`
	Meta          map[string]any `json:"meta"`
}

type ListDBBackupsResponse struct {
	Items []DBBackupItem `json:"items"`
}

type PresignDBBackupResponse struct {
	BackupID        string `json:"backup_id"`
	URL             string `json:"url"`
	ExpiresInSeconds int64 `json:"expires_in_seconds"`
}
