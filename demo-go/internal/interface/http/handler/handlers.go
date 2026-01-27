package handler

import (
	"dag-observatory/demo-go/internal/application/port"
	"dag-observatory/demo-go/internal/infrastructure/observability/metrics"

	"go.opentelemetry.io/otel/trace"
)

type Handlers struct {
	intentLog port.IntentLog
	tracer    trace.Tracer
	metrics   *metrics.Instruments
}

type Dependencies struct {
	IntentLog port.IntentLog
	Tracer    trace.Tracer
	Metrics   *metrics.Instruments
}

func New(d Dependencies) *Handlers {
	return &Handlers{
		intentLog: d.IntentLog,
		tracer:    d.Tracer,
		metrics:   d.Metrics,
	}
}
