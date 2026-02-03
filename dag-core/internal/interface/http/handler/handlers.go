package handler

import (
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/application/observability/port"
	"dag-observatory/dag-core/internal/infrastructure/observability/applog"
	"dag-observatory/dag-core/internal/infrastructure/observability/metrics"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	miniostore "dag-observatory/dag-core/internal/infrastructure/storage/minio"

	"go.opentelemetry.io/otel/trace"
)

type Handlers struct {
	intentLog port.IntentLog
	appLog    *applog.Logger
	tracer    trace.Tracer
	metrics   *metrics.Instruments
	runWF     *usecase.RunWorkflow
	artifacts *tsdb.ArtifactsRepository
	presigner *miniostore.Presigner
	dbBackups *tsdb.DBBackupsRepository
}

type Dependencies struct {
	IntentLog  port.IntentLog
	AppLog     *applog.Logger
	Tracer     trace.Tracer
	Metrics    *metrics.Instruments
	RunWorkflow *usecase.RunWorkflow
	ArtifactsRepo *tsdb.ArtifactsRepository
	Presigner *miniostore.Presigner
	DBBackupsRepo *tsdb.DBBackupsRepository
}

func New(d Dependencies) *Handlers {
	return &Handlers{
		intentLog: d.IntentLog,
		appLog:    d.AppLog,
		tracer:    d.Tracer,
		metrics:   d.Metrics,
		runWF:     d.RunWorkflow,
		artifacts: d.ArtifactsRepo,
		presigner: d.Presigner,
		dbBackups: d.DBBackupsRepo,
	}
}
