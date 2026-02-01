package handler

import (
	"log/slog"
	"net/http"
	"time"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/observability/ctxprop"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type dagRunReq struct {
	Symbol string `json:"symbol"`
	Mode   string `json:"mode"`
}

func (h *Handlers) DagRun(c echo.Context) error {
	ctx := c.Request().Context()

	if h.runWF == nil {
		h.appLog.Error(ctx, "dag runtime usecase not configured")
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "dag runtime not configured",
		})
	}

	var req dagRunReq
	if err := c.Bind(&req); err != nil {
		h.appLog.Warn(ctx, "dag run request bind failed",
			slog.String("error", err.Error()),
		)
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "invalid request",
		})
	}
	if req.Symbol == "" {
		req.Symbol = "USDJPY"
	}
	if req.Mode == "" {
		req.Mode = c.QueryParam("mode")
	}
	if req.Mode == "" {
		req.Mode = "normal"
	}

	runID := uuid.NewString()
	ctx = ctxprop.WithRunID(ctx, runID)

	partition := state.Partition(runID)
	payload := map[string]any{
		"run_id":    runID,
		"task_id":   "heavy_calc",
		"attempt":   1,
		"task_name": "heavy_calc",
		"input": map[string]any{
			"mode":   req.Mode,
			"symbol": req.Symbol,
		},
		// Keep top-level keys for direct execution compatibility.
		"mode":   req.Mode,
		"symbol": req.Symbol,
	}

	event := events.Event{
		EventID:   runID,
		EventTime: time.Now(),
		Partition: partition,
		Type:      "task.requested",
		Payload:   payload,
	}

	if err := h.runWF.Handle(ctx, partition, event); err != nil {
		h.appLog.Error(ctx, "dag run failed",
			slog.String("error", err.Error()),
		)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"status": "failed",
			"symbol": req.Symbol,
			"run_id": runID,
		})
	}

	if h.runWF.IsEnqueueMode() {
		return c.JSON(http.StatusAccepted, map[string]any{
			"status": "enqueued",
			"symbol": req.Symbol,
			"run_id": runID,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "completed",
		"symbol": req.Symbol,
		"run_id": runID,
	})
}
