package di

import (
	"dag-observatory/demo-go/internal/infrastructure/observability/applog"
	httpif "dag-observatory/demo-go/internal/interface/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func NewEchoContainer(cfg Config, otelc *OTelContainer) (*echo.Echo, error) {
	e := echo.New()

	// OTel middleware (official instrumentation)
	e.Use(otelecho.Middleware(cfg.ServiceName))

	appLog := applog.New(cfg.AppLogLevel, cfg.AppLogOutput, otelc.AppLogger)

	httpif.RegisterRoutes(e, httpif.Dependencies{
		IntentLog: otelc.IntentLog,
		AppLog:    appLog,
		Tracer:    otelc.Tracer,
		Metrics:   otelc.Metrics,
	})

	return e, nil
}
