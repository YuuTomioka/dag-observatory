package http

import (
	artifactsrepository "dag-observatory/dag-core/internal/application/artifacts/repository"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	dbbackupsrepository "dag-observatory/dag-core/internal/application/dbbackups/repository"
	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/application/observability/port"
	"dag-observatory/dag-core/internal/infrastructure/observability/applog"
	"dag-observatory/dag-core/internal/infrastructure/observability/metrics"
	miniostore "dag-observatory/dag-core/internal/infrastructure/storage/minio"
	"dag-observatory/dag-core/internal/interface/http/handler"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
)

type Dependencies struct {
	IntentLog     port.IntentLog
	AppLog        *applog.Logger
	Tracer        trace.Tracer
	Metrics       *metrics.Instruments
	RunWorkflow   *usecase.RunWorkflow
	ArtifactsRepo artifactsrepository.Reader
	Presigner     *miniostore.Presigner
	DBBackupsRepo dbbackupsrepository.Reader

	CreateSymbol              *marketdatausecase.CreateSymbol
	GetSymbolByCode           *marketdatausecase.GetSymbolByCode
	ListSymbols               *marketdatausecase.ListSymbols
	UpsertTicks               *marketdatausecase.UpsertTicks
	GetLatestTickBySymbol     *marketdatausecase.GetLatestTickBySymbol
	ListTicksBySymbolAndRange *marketdatausecase.ListTicksBySymbolAndRange
}

func RegisterRoutes(e *echo.Echo, d Dependencies) {
	h := handler.New(handler.Dependencies{
		IntentLog:                 d.IntentLog,
		AppLog:                    d.AppLog,
		Tracer:                    d.Tracer,
		Metrics:                   d.Metrics,
		RunWorkflow:               d.RunWorkflow,
		ArtifactsRepo:             d.ArtifactsRepo,
		Presigner:                 d.Presigner,
		DBBackupsRepo:             d.DBBackupsRepo,
		CreateSymbol:              d.CreateSymbol,
		GetSymbolByCode:           d.GetSymbolByCode,
		ListSymbols:               d.ListSymbols,
		UpsertTicks:               d.UpsertTicks,
		GetLatestTickBySymbol:     d.GetLatestTickBySymbol,
		ListTicksBySymbolAndRange: d.ListTicksBySymbolAndRange,
	})

	e.GET("/healthz", h.Healthz)
	e.POST("/dag/run", h.DagRun)
	e.GET("/workflow-runs/:workflow_run_id/artifacts", h.ListArtifactsByWorkflow)
	e.POST("/artifacts/:artifact_id:presign-download", h.PresignArtifact)
	e.GET("/db-backups", h.ListDBBackups)
	e.POST("/db-backups/:backup_id:presign-download", h.PresignDBBackup)
	e.POST("/marketdata/symbols", h.CreateSymbol)
	e.GET("/marketdata/symbols", h.ListSymbols)
	e.GET("/marketdata/symbols/:code", h.GetSymbolByCode)
	e.POST("/marketdata/ticks:upsert-bulk", h.UpsertTicksBulk)
	e.GET("/marketdata/ticks/latest", h.GetLatestTickBySymbol)
	e.GET("/marketdata/ticks", h.ListTicksBySymbolAndRange)
}
