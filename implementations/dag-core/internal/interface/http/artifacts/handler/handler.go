package handler

import (
	artifactsrepository "dag-observatory/dag-core/internal/application/artifacts/repository"
	miniostore "dag-observatory/dag-core/internal/infrastructure/storage/minio"
)

type Handler struct {
	artifacts artifactsrepository.Reader
	presigner *miniostore.Presigner
}

type Dependencies struct {
	ArtifactsRepo artifactsrepository.Reader
	Presigner     *miniostore.Presigner
}

func New(d Dependencies) *Handler {
	return &Handler{
		artifacts: d.ArtifactsRepo,
		presigner: d.Presigner,
	}
}
