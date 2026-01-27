package http

import (
	"dag-observatory/demo-go/internal/application/port"
	"dag-observatory/demo-go/internal/interface/http/handler"
	"dag-observatory/demo-go/internal/infrastructure/observability/metrics"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
)

type Dependencies struct {
	IntentLog port.IntentLog
	Tracer    trace.Tracer
	Metrics   *metrics.Instruments
}

func RegisterRoutes(e *echo.Echo, d Dependencies) {
	h := handler.New(handler.Dependencies{
		IntentLog: d.IntentLog,
		Tracer:    d.Tracer,
		Metrics:   d.Metrics,
	})

	e.GET("/healthz", h.Healthz)
	e.POST("/dag/run", h.DagRun)
}
