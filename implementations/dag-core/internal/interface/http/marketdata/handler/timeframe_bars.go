package handler

import (
	"net/http"

	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/interface/http/dto"
	"dag-observatory/dag-core/internal/interface/http/marketdata/request"
	"dag-observatory/dag-core/internal/interface/http/marketdata/response"

	"github.com/labstack/echo/v4"
)

// BackfillTimeframeBars
// @Summary Backfill timeframe bars directly
// @Description Rebuilds timeframe bars immediately and returns aggregated bar counts for the requested symbol and range. Use this when the caller needs the backfill result itself.
// @Tags marketdata
// @Accept json
// @Produce json
// @Param request body request.BackfillTimeframeBarsRequest true "Backfill timeframe bars request"
// @Success 200 {object} response.BackfillTimeframeBarsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /marketdata/timeframe-bars:backfill [post]
func (h *Handlers) BackfillTimeframeBars(c echo.Context) error {
	if h.backfillTimeframeBars == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "marketdata usecases not configured",
			Status: "error",
		})
	}

	var req request.BackfillTimeframeBarsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "invalid request",
			Status: "error",
		})
	}

	result, err := h.backfillTimeframeBars.Execute(c.Request().Context(), marketdatausecase.BackfillTimeframeBarsRequest{
		SymbolID:      marketdata.SymbolID(req.SymbolID),
		SymbolCode:    req.SymbolCode,
		TimeframeCode: marketdata.TimeframeCode(req.TimeframeCode),
		From:          req.From,
		To:            req.To,
		Mode:          marketdatausecase.BackfillMode(req.Mode),
		ChunkSizeBars: req.ChunkSizeBars,
	})
	if err != nil {
		return writeMarketDataError(c, err)
	}

	return c.JSON(http.StatusOK, response.BackfillTimeframeBarsResponse{
		Status:        "completed",
		Entrypoint:    "direct",
		SymbolID:      int64(result.SymbolID),
		TimeframeCode: string(result.TimeframeCode),
		From:          result.From.String(),
		To:            result.To.String(),
		ChunkCount:    result.ChunkCount,
		BarCount:      result.BarCount,
	})
}

// BackfillTimeframeBarsWorkflow
// @Summary Trigger timeframe bar backfill workflow
// @Description Enqueues or runs the configured marketdata backfill workflow and returns workflow run information instead of backfill aggregation counts. Use this when the caller wants workflow execution semantics under the marketdata feature.
// @Tags marketdata
// @Accept json
// @Produce json
// @Param request body request.BackfillTimeframeBarsRequest true "Backfill timeframe bars workflow request"
// @Success 200 {object} response.BackfillTimeframeBarsWorkflowResponse
// @Success 202 {object} response.BackfillTimeframeBarsWorkflowResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /marketdata/timeframe-bars:backfill-workflow [post]
func (h *Handlers) BackfillTimeframeBarsWorkflow(c echo.Context) error {
	if h.runWorkflow == nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  "dag runtime not configured",
			Status: "error",
		})
	}

	var req request.BackfillTimeframeBarsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "invalid request",
			Status: "error",
		})
	}

	result, err := h.runWorkflow.Execute(c.Request().Context(), dagruntimeusecase.RunWorkflowRequest{
		Marketdata: &dagruntimeusecase.MarketdataRunInput{
			SymbolID:      req.SymbolID,
			SymbolCode:    req.SymbolCode,
			TimeframeCode: req.TimeframeCode,
			From:          req.From,
			To:            req.To,
		},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  "workflow run failed",
			Status: "failed",
			Symbol: result.Symbol,
			RunID:  result.RunID,
		})
	}

	resp := response.BackfillTimeframeBarsWorkflowResponse{
		Status:     "completed",
		Entrypoint: "workflow",
		Symbol:     result.Symbol,
		RunID:      result.RunID,
		Marketdata: mapBackfillWorkflowInput(result.Marketdata),
	}
	if result.EnqueueMode {
		resp.Status = "enqueued"
		return c.JSON(http.StatusAccepted, resp)
	}
	return c.JSON(http.StatusOK, resp)
}

func mapBackfillWorkflowInput(input *dagruntimeusecase.MarketdataRunInput) *response.BackfillWorkflowInput {
	if input == nil {
		return nil
	}
	return &response.BackfillWorkflowInput{
		SymbolID:      input.SymbolID,
		SymbolCode:    input.SymbolCode,
		TimeframeCode: input.TimeframeCode,
		From:          input.From.String(),
		To:            input.To.String(),
	}
}
