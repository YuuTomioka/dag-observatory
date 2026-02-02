package handler

import (
	"log/slog"
	"net/http"

	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/observability/semantics"
	"dag-observatory/dag-core/internal/interface/http/dto"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DagRun
// @Summary Run workflow
// @Description Triggers workflow execution (sync/async depending on backend)
// @Tags dag
// @Accept json
// @Produce json
// @Param request body dto.DagRunRequest true "Run workflow request"
// @Success 200 {object} map[string]any
// @Success 202 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /dag/run [post]
func (h *Handlers) DagRun(c echo.Context) error {
	ctx := c.Request().Context()

	if h.runWF == nil {
		h.appLog.Error(ctx, "dag runtime usecase not configured")
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "dag runtime not configured",
		})
	}

	var req dto.DagRunRequest
	if err := c.Bind(&req); err != nil {
		h.appLog.Warn(ctx, "dag run request bind failed",
			slog.String("error", err.Error()),
		)
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "invalid request",
		})
	}
	mode := req.Mode
	if mode == "" {
		mode = c.QueryParam("mode")
	}

	result, err := h.runWF.Execute(ctx, usecase.RunWorkflowRequest{
		Symbol: req.Symbol,
		Mode:   mode,
	})
	if err != nil {
		h.appLog.Error(ctx, "dag run failed",
			slog.String("error", err.Error()),
		)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"status": "failed",
			"symbol": result.Symbol,
			"run_id": result.RunID,
		})
	}

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String(semantics.KeyDAGRunID, result.RunID),
		attribute.String(semantics.KeyDAGTaskID, result.TaskID),
		attribute.String(semantics.KeyDAGTaskName, result.TaskName),
		attribute.Int(semantics.KeyDAGAttempt, result.Attempt),
		attribute.String(semantics.KeyMessagingSystem, "kafka"),
		attribute.String(semantics.KeyMessagingDestination, "dagruntime-events"),
		attribute.String(semantics.KeyMessagingOperation, "send"),
		attribute.String(semantics.KeyMessagingMessageID, result.RunID),
	)

	if result.EnqueueMode {
		return c.JSON(http.StatusAccepted, map[string]any{
			"status": "enqueued",
			"symbol": result.Symbol,
			"run_id": result.RunID,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "completed",
		"symbol": result.Symbol,
		"run_id": result.RunID,
	})
}
