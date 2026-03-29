package repository

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("artifacts repository: not found")
	ErrNotConfigured = errors.New("artifacts repository: not configured")
)

type Artifact struct {
	ArtifactID     string
	WorkflowRunID  string
	TaskID         string
	AttemptNo      int
	Bucket         string
	ObjectKey      string
	Kind           string
	Filename       string
	ContentType    string
	SizeBytes      int64
	ChecksumSHA256 string
	ETag           string
	Status         string
	CreatedAt      time.Time
	Tags           map[string]any
}

type Reader interface {
	ListByWorkflow(ctx context.Context, workflowRunID string, limit, offset int) ([]Artifact, error)
	GetByID(ctx context.Context, artifactID string) (Artifact, error)
}
