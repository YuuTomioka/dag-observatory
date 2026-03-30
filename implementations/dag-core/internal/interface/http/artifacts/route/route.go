package route

import (
	artifactshandler "dag-observatory/dag-core/internal/interface/http/artifacts/handler"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, h *artifactshandler.Handler) {
	e.GET("/workflow-runs/:workflow_run_id/artifacts", h.ListArtifactsByWorkflow)
	e.POST("/artifacts/:artifact_id\\:presign-download", h.PresignArtifact)
}
