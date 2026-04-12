package route

import (
	daghandler "dag-observatory/dag-core/internal/interface/http/dag/handler"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, h *daghandler.Handler) {
	e.POST("/dag/run", h.DagRun)
	e.GET("/runs", h.ListRuns)
	e.GET("/runs/ui", h.RunsUI)
	e.GET("/runs/:run_id", h.GetRun)
	e.GET("/runs/:run_id/steps", h.ListRunSteps)
	e.GET("/runs/:run_id/steps/:sequence_no", h.GetRunNode)
	// Deprecated: compatibility alias for clients that still use execution_id path naming.
	e.GET("/runs/:run_id/nodes/:execution_id", h.GetRunNode)
}
