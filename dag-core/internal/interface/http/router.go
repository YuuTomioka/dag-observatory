package http

import (
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/application/observability/port"
	"dag-observatory/dag-core/internal/infrastructure/observability/applog"
	"dag-observatory/dag-core/internal/infrastructure/observability/metrics"
	"dag-observatory/dag-core/internal/interface/http/handler"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
)

type Dependencies struct {
	IntentLog   port.IntentLog
	AppLog      *applog.Logger
	Tracer      trace.Tracer
	Metrics     *metrics.Instruments
	RunWorkflow *usecase.RunWorkflow
}

func RegisterRoutes(e *echo.Echo, d Dependencies) {
	h := handler.New(handler.Dependencies{
		IntentLog:  d.IntentLog,
		AppLog:     d.AppLog,
		Tracer:     d.Tracer,
		Metrics:    d.Metrics,
		RunWorkflow: d.RunWorkflow,
	})

	e.GET("/healthz", h.Healthz)
	e.POST("/dag/run", h.DagRun)
}
