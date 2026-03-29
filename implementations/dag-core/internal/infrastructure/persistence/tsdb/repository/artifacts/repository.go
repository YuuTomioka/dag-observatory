package artifacts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	apprepository "dag-observatory/dag-core/internal/application/artifacts/repository"
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

func (r *Repository) ListByWorkflow(ctx context.Context, workflowRunID string, limit, offset int) ([]apprepository.Artifact, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListArtifactsByWorkflow(ctx, query.ListArtifactsByWorkflowParams{
		WorkflowRunID: workflowRunID,
		Limit:         int32(limit),
		Offset:        int32(offset),
	})
	if err != nil {
		return nil, mapError(err)
	}

	out := make([]apprepository.Artifact, 0, len(rows))
	for _, row := range rows {
		mapped, err := fromListRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}
	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, artifactID string) (apprepository.Artifact, error) {
	if err := r.validate(); err != nil {
		return apprepository.Artifact{}, err
	}

	parsedID, err := uuid.Parse(artifactID)
	if err != nil {
		return apprepository.Artifact{}, fmt.Errorf("artifacts repository: invalid artifact_id: %w", err)
	}
	row, err := r.queries.GetArtifactByID(ctx, toPgUUID(parsedID))
	if err != nil {
		return apprepository.Artifact{}, mapError(err)
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

func fromListRow(row query.ListArtifactsByWorkflowRow) (apprepository.Artifact, error) {
	id, err := fromPgUUID(row.ArtifactID)
	if err != nil {
		return apprepository.Artifact{}, err
	}
	createdAt, err := fromPgTime(row.CreatedAt)
	if err != nil {
		return apprepository.Artifact{}, err
	}
	tags, err := parseJSONMap(row.Tags)
	if err != nil {
		return apprepository.Artifact{}, err
	}

	return apprepository.Artifact{
		ArtifactID:    id.String(),
		WorkflowRunID: row.WorkflowRunID,
		TaskID:        row.TaskID,
		AttemptNo:     int(row.AttemptNo),
		Kind:          row.Kind,
		Filename:      row.Filename.String,
		ContentType:   row.ContentType.String,
		SizeBytes:     row.SizeBytes,
		Status:        row.Status,
		CreatedAt:     createdAt,
		Tags:          tags,
	}, nil
}

func fromRow(row query.Artifact) (apprepository.Artifact, error) {
	id, err := fromPgUUID(row.ArtifactID)
	if err != nil {
		return apprepository.Artifact{}, err
	}
	createdAt, err := fromPgTime(row.CreatedAt)
	if err != nil {
		return apprepository.Artifact{}, err
	}
	tags, err := parseJSONMap(row.Tags)
	if err != nil {
		return apprepository.Artifact{}, err
	}

	return apprepository.Artifact{
		ArtifactID:     id.String(),
		WorkflowRunID:  row.WorkflowRunID,
		TaskID:         row.TaskID,
		AttemptNo:      int(row.AttemptNo),
		Bucket:         row.Bucket,
		ObjectKey:      row.ObjectKey,
		Kind:           row.Kind,
		Filename:       row.Filename.String,
		ContentType:    row.ContentType.String,
		SizeBytes:      row.SizeBytes,
		ChecksumSHA256: row.ChecksumSha256.String,
		ETag:           row.Etag.String,
		Status:         row.Status,
		CreatedAt:      createdAt,
		Tags:           tags,
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
		return uuid.UUID{}, fmt.Errorf("artifacts repository: uuid is NULL")
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
		return time.Time{}, fmt.Errorf("artifacts repository: timestamp is NULL")
	}
	return v.Time, nil
}

var _ apprepository.Reader = (*Repository)(nil)
