package handler

import (
	"net/http"

	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/interface/http/dto"
	"dag-observatory/dag-core/internal/interface/http/marketdata/request"
	"dag-observatory/dag-core/internal/interface/http/marketdata/response"

	"github.com/labstack/echo/v4"
)

// BackfillTimeframeBars
// @Summary Backfill timeframe bars
// @Description Rebuilds timeframe bars from ticks for a symbol and range
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
		SymbolID:      int64(result.SymbolID),
		TimeframeCode: string(result.TimeframeCode),
		From:          result.From.String(),
		To:            result.To.String(),
		ChunkCount:    result.ChunkCount,
		BarCount:      result.BarCount,
	})
}
