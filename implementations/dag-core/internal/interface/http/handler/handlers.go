package handler

import (
	artifactsrepository "dag-observatory/dag-core/internal/application/artifacts/repository"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	dbbackupsrepository "dag-observatory/dag-core/internal/application/dbbackups/repository"
	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/application/observability/port"
	"dag-observatory/dag-core/internal/infrastructure/observability/applog"
	"dag-observatory/dag-core/internal/infrastructure/observability/metrics"
	miniostore "dag-observatory/dag-core/internal/infrastructure/storage/minio"

	"go.opentelemetry.io/otel/trace"
)

type Handlers struct {
	intentLog port.IntentLog
	appLog    *applog.Logger
	tracer    trace.Tracer
	metrics   *metrics.Instruments
	runWF     *usecase.RunWorkflow
	artifacts artifactsrepository.Reader
	presigner *miniostore.Presigner
	dbBackups dbbackupsrepository.Reader

	createSymbol              *marketdatausecase.CreateSymbol
	getSymbolByCode           *marketdatausecase.GetSymbolByCode
	listSymbols               *marketdatausecase.ListSymbols
	upsertTicks               *marketdatausecase.UpsertTicks
	getLatestTickBySymbol     *marketdatausecase.GetLatestTickBySymbol
	listTicksBySymbolAndRange *marketdatausecase.ListTicksBySymbolAndRange
}

type Dependencies struct {
	IntentLog                 port.IntentLog
	AppLog                    *applog.Logger
	Tracer                    trace.Tracer
	Metrics                   *metrics.Instruments
	RunWorkflow               *usecase.RunWorkflow
	ArtifactsRepo             artifactsrepository.Reader
	Presigner                 *miniostore.Presigner
	DBBackupsRepo             dbbackupsrepository.Reader
	CreateSymbol              *marketdatausecase.CreateSymbol
	GetSymbolByCode           *marketdatausecase.GetSymbolByCode
	ListSymbols               *marketdatausecase.ListSymbols
	UpsertTicks               *marketdatausecase.UpsertTicks
	GetLatestTickBySymbol     *marketdatausecase.GetLatestTickBySymbol
	ListTicksBySymbolAndRange *marketdatausecase.ListTicksBySymbolAndRange
}

func New(d Dependencies) *Handlers {
	return &Handlers{
		intentLog:                 d.IntentLog,
		appLog:                    d.AppLog,
		tracer:                    d.Tracer,
		metrics:                   d.Metrics,
		runWF:                     d.RunWorkflow,
		artifacts:                 d.ArtifactsRepo,
		presigner:                 d.Presigner,
		dbBackups:                 d.DBBackupsRepo,
		createSymbol:              d.CreateSymbol,
		getSymbolByCode:           d.GetSymbolByCode,
		listSymbols:               d.ListSymbols,
		upsertTicks:               d.UpsertTicks,
		getLatestTickBySymbol:     d.GetLatestTickBySymbol,
		listTicksBySymbolAndRange: d.ListTicksBySymbolAndRange,
	}
}
