package route

import (
	daghandler "dag-observatory/dag-core/internal/interface/http/dag/handler"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, h *daghandler.Handler) {
	e.POST("/dag/run", h.DagRun)
}
