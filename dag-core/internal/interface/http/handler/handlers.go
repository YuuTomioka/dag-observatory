package handler

import (
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/application/port"
	"dag-observatory/dag-core/internal/infrastructure/observability/applog"
	"dag-observatory/dag-core/internal/infrastructure/observability/metrics"

	"go.opentelemetry.io/otel/trace"
)

type Handlers struct {
	intentLog port.IntentLog
	appLog    *applog.Logger
	tracer    trace.Tracer
	metrics   *metrics.Instruments
	runWF     *usecase.RunWorkflow
}

type Dependencies struct {
	IntentLog  port.IntentLog
	AppLog     *applog.Logger
	Tracer     trace.Tracer
	Metrics    *metrics.Instruments
	RunWorkflow *usecase.RunWorkflow
}

func New(d Dependencies) *Handlers {
	return &Handlers{
		intentLog: d.IntentLog,
		appLog:    d.AppLog,
		tracer:    d.Tracer,
		metrics:   d.Metrics,
		runWF:     d.RunWorkflow,
	}
}
