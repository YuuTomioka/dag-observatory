package route

import (
	healthhandler "dag-observatory/dag-core/internal/interface/http/health/handler"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, h *healthhandler.Handler) {
	e.GET("/healthz", h.Healthz)
}
