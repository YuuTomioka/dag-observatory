package di

import (
	"dag-observatory/demo-go/internal/infrastructure/observability/otelecho"
	httpif "dag-observatory/demo-go/internal/interface/http"

	"github.com/labstack/echo/v4"
)

func NewEchoContainer(cfg Config, otelc *OTelContainer) (*echo.Echo, error) {
	e := echo.New()

	// OTel middleware (trace context)
	e.Use(otelecho.Middleware(otelc.Tracer))

	httpif.RegisterRoutes(e, httpif.Dependencies{
		IntentLog: otelc.IntentLog,
		Tracer:    otelc.Tracer,
		Metrics:   otelc.Metrics,
	})

	return e, nil
}
