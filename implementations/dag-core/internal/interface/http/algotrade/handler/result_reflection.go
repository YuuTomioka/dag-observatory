package handler

import (
	"net/http"

	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/marketdata/ohlc"
	"dag-observatory/dag-core/internal/interface/http/algotrade/response"
	"dag-observatory/dag-core/internal/interface/http/dto"

	"github.com/labstack/echo/v4"
)

type resultReflectionRunRequest struct {
	Partition      string       `json:"partition"`
	RunID          string       `json:"run_id,omitempty"`
	Symbol         string       `json:"symbol,omitempty"`
	Mode           string       `json:"mode,omitempty"`
	OHLCVBars      []ohlc.OHLCV `json:"ohlcv_bars"`
	SpreadBps      float64      `json:"spread_bps,omitempty"`
	AccountBalance float64      `json:"account_balance,omitempty"`
}

// RunResultReflection
// @Summary Run result reflection workflow
// @Description Triggers the configured runtime workflow as a result-reflection entrypoint. Intended for manual submit/fill/result-reflection verification with explicit partition control.
// @Tags algotrade
// @Accept json
// @Produce json
// @Param request body resultReflectionRunRequest true "Result reflection run request"
// @Success 200 {object} response.ResultReflectionRunResponse
// @Success 202 {object} response.ResultReflectionRunResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /algotrade/result-reflection:run [post]
func (h *Handler) RunResultReflection(c echo.Context) error {
	if h.runWF == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "result reflection API not configured",
			Status: "error",
		})
	}

	var req resultReflectionRunRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "invalid request",
			Status: "error",
		})
	}
	if req.Partition == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "partition is required",
			Status: "error",
		})
	}
	if len(req.OHLCVBars) == 0 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "ohlcv_bars is required",
			Status: "error",
		})
	}

	result, err := h.runWF.Execute(c.Request().Context(), dagruntimeusecase.RunWorkflowRequest{
		RunID:          req.RunID,
		Partition:      req.Partition,
		Symbol:         req.Symbol,
		Mode:           req.Mode,
		OHLCVBars:      req.OHLCVBars,
		SpreadBps:      req.SpreadBps,
		AccountBalance: req.AccountBalance,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "failed",
			Symbol: result.Symbol,
			RunID:  result.RunID,
		})
	}

	resp := response.ResultReflectionRunResponse{
		Entrypoint: "result_reflection",
		Partition:  req.Partition,
		RunID:      result.RunID,
		Symbol:     result.Symbol,
	}
	if result.EnqueueMode {
		resp.Status = "enqueued"
		return c.JSON(http.StatusAccepted, resp)
	}
	resp.Status = "completed"
	return c.JSON(http.StatusOK, resp)
}
