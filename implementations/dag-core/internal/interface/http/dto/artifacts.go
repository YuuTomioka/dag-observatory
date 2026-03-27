package dto

import "time"

type ArtifactItem struct {
	ArtifactID    string         `json:"artifact_id"`
	WorkflowRunID string         `json:"workflow_run_id"`
	TaskID        string         `json:"task_id"`
	AttemptNo     int            `json:"attempt_no"`
	Kind          string         `json:"kind"`
	Filename      string         `json:"filename,omitempty"`
	ContentType   string         `json:"content_type,omitempty"`
	SizeBytes     int64          `json:"size_bytes"`
	Status        string         `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	Tags          map[string]any `json:"tags"`
}

type ListArtifactsResponse struct {
	Items []ArtifactItem `json:"items"`
}

type PresignResponse struct {
	ArtifactID       string `json:"artifact_id"`
	URL              string `json:"url"`
	ExpiresInSeconds int64  `json:"expires_in_seconds"`
}
