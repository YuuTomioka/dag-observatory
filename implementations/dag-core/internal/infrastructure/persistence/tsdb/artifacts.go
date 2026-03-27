package tsdb

import (
	"context"
	"fmt"
	"time"
)

type Artifact struct {
	ArtifactID    string
	WorkflowRunID string
	TaskID        string
	AttemptNo     int
	Bucket        string
	ObjectKey     string
	Kind          string
	Filename      string
	ContentType   string
	SizeBytes     int64
	ChecksumSHA256 string
	ETag           string
	Status         string
	CreatedAt      time.Time
	Tags           map[string]any
}

type ArtifactsRepository struct {
	pool *Client
}

func NewArtifactsRepository(client *Client) *ArtifactsRepository {
	return &ArtifactsRepository{pool: client}
}

func (r *ArtifactsRepository) ListByWorkflow(ctx context.Context, workflowRunID string, limit, offset int) ([]Artifact, error) {
	if r == nil || r.pool == nil || r.pool.Pool == nil {
		return nil, fmt.Errorf("tsdb: client not configured")
	}
	rows, err := r.pool.Pool.Query(ctx, `
SELECT artifact_id, workflow_run_id, task_id, attempt_no, bucket, object_key, kind,
       COALESCE(filename,''), COALESCE(content_type,''), size_bytes,
       COALESCE(checksum_sha256,''), COALESCE(etag,''), status, created_at, tags
FROM artifacts
WHERE workflow_run_id=$1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`, workflowRunID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Artifact, 0, limit)
	for rows.Next() {
		var a Artifact
		if err := rows.Scan(
			&a.ArtifactID,
			&a.WorkflowRunID,
			&a.TaskID,
			&a.AttemptNo,
			&a.Bucket,
			&a.ObjectKey,
			&a.Kind,
			&a.Filename,
			&a.ContentType,
			&a.SizeBytes,
			&a.ChecksumSHA256,
			&a.ETag,
			&a.Status,
			&a.CreatedAt,
			&a.Tags,
		); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *ArtifactsRepository) GetByID(ctx context.Context, artifactID string) (Artifact, error) {
	if r == nil || r.pool == nil || r.pool.Pool == nil {
		return Artifact{}, fmt.Errorf("tsdb: client not configured")
	}
	var a Artifact
	err := r.pool.Pool.QueryRow(ctx, `
SELECT artifact_id, workflow_run_id, task_id, attempt_no, bucket, object_key, kind,
       COALESCE(filename,''), COALESCE(content_type,''), size_bytes,
       COALESCE(checksum_sha256,''), COALESCE(etag,''), status, created_at, tags
FROM artifacts
WHERE artifact_id=$1
`, artifactID).Scan(
		&a.ArtifactID,
		&a.WorkflowRunID,
		&a.TaskID,
		&a.AttemptNo,
		&a.Bucket,
		&a.ObjectKey,
		&a.Kind,
		&a.Filename,
		&a.ContentType,
		&a.SizeBytes,
		&a.ChecksumSHA256,
		&a.ETag,
		&a.Status,
		&a.CreatedAt,
		&a.Tags,
	)
	return a, err
}
