package handler

import (
	"log/slog"
	"net/http"

	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/observability/semantics"
	"dag-observatory/dag-core/internal/interface/http/dag/request"
	"dag-observatory/dag-core/internal/interface/http/dag/response"
	"dag-observatory/dag-core/internal/interface/http/dto"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DagRun
// @Summary Run workflow generically
// @Description Triggers the configured workflow using generic runtime input. Prefer feature-local workflow endpoints when a dedicated endpoint exists and use this endpoint for runtime-level execution.
// @Tags dag
// @Accept json
// @Produce json
// @Param request body request.RunRequest true "Run workflow request"
// @Param mode query string false "Override mode when body omits it"
// @Success 200 {object} response.RunResponse
// @Success 202 {object} response.RunResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /dag/run [post]
func (h *Handler) DagRun(c echo.Context) error {
	ctx := c.Request().Context()

	if h.runWF == nil {
		h.appLog.Error(ctx, "dag runtime usecase not configured")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "dag runtime not configured",
		})
	}

	var req request.RunRequest
	if err := c.Bind(&req); err != nil {
		h.appLog.Warn(ctx, "dag run request bind failed",
			slog.String("error", err.Error()),
		)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "invalid request",
		})
	}
	mode := req.Mode
	if mode == "" {
		mode = c.QueryParam("mode")
	}

	result, err := h.runWF.Execute(ctx, usecase.RunWorkflowRequest{
		Symbol:     req.Symbol,
		Mode:       mode,
		Bars:       req.Bars,
		Marketdata: mapMarketdataInput(req.Marketdata),
	})
	if err != nil {
		h.appLog.Error(ctx, "dag run failed",
			slog.String("error", err.Error()),
		)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Status: "failed",
			Symbol: result.Symbol,
			RunID:  result.RunID,
		})
	}

	span := trace.SpanFromContext(ctx)
	attrs := []attribute.KeyValue{
		attribute.String(semantics.KeyDAGRunID, result.RunID),
		attribute.String(semantics.KeyDAGTaskID, result.TaskID),
		attribute.String(semantics.KeyDAGTaskName, result.TaskName),
		attribute.Int(semantics.KeyDAGAttempt, result.Attempt),
		attribute.String(semantics.KeyMessagingDestination, result.MessagingDestination),
		attribute.String(semantics.KeyMessagingOperation, "send"),
		attribute.String(semantics.KeyMessagingMessageID, result.RunID),
	}
	if result.EnqueueMode {
		attrs = append(attrs, attribute.String(semantics.KeyMessagingSystem, "kafka"))
	} else {
		attrs = append(attrs, attribute.String(semantics.KeyMessagingSystem, "inproc"))
	}
	span.SetAttributes(attrs...)

	if result.EnqueueMode {
		return c.JSON(http.StatusAccepted, response.RunResponse{
			Status:     "enqueued",
			Entrypoint: "workflow",
			Symbol:     result.Symbol,
			RunID:      result.RunID,
			Marketdata: mapMarketdataResponse(result.Marketdata),
		})
	}

	return c.JSON(http.StatusOK, response.RunResponse{
		Status:     "completed",
		Entrypoint: "workflow",
		Symbol:     result.Symbol,
		RunID:      result.RunID,
		Marketdata: mapMarketdataResponse(result.Marketdata),
	})
}

func mapMarketdataInput(input *request.MarketdataInput) *usecase.MarketdataRunInput {
	if input == nil {
		return nil
	}
	return &usecase.MarketdataRunInput{
		SymbolID:      input.SymbolID,
		SymbolCode:    input.SymbolCode,
		TimeframeCode: input.TimeframeCode,
		From:          input.From,
		To:            input.To,
	}
}

func mapMarketdataResponse(input *usecase.MarketdataRunInput) *response.MarketdataInput {
	if input == nil {
		return nil
	}
	return &response.MarketdataInput{
		SymbolID:      input.SymbolID,
		SymbolCode:    input.SymbolCode,
		TimeframeCode: input.TimeframeCode,
		From:          input.From.String(),
		To:            input.To.String(),
	}
}
