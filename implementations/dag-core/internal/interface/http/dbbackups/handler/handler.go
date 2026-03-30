package handler

import (
	dbbackupsrepository "dag-observatory/dag-core/internal/application/dbbackups/repository"
	miniostore "dag-observatory/dag-core/internal/infrastructure/storage/minio"
)

type Handler struct {
	dbBackups dbbackupsrepository.Reader
	presigner *miniostore.Presigner
}

type Dependencies struct {
	DBBackupsRepo dbbackupsrepository.Reader
	Presigner     *miniostore.Presigner
}

func New(d Dependencies) *Handler {
	return &Handler{
		dbBackups: d.DBBackupsRepo,
		presigner: d.Presigner,
	}
}
