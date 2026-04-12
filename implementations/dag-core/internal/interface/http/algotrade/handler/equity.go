package handler

import (
	"net/http"
	"strings"

	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/algotrade"
	domainstate "dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/interface/http/algotrade/response"
	"dag-observatory/dag-core/internal/interface/http/dto"

	"github.com/labstack/echo/v4"
)

// GetEquitySeries
// @Summary Get equity/drawdown time series
// @Description Returns equity and drawdown points reconstructed from closed trades in the selected partition or run scope.
// @Tags algotrade
// @Produce json
// @Param partition query string false "Runtime partition"
// @Param run_id query string false "Run ID used to resolve partition scope when partition is omitted"
// @Param symbol query string false "Symbol filter"
// @Param side query string false "Position side filter (long/short)"
// @Param from query string false "Exit-time lower bound (RFC3339 or RFC3339Nano, inclusive)"
// @Param to query string false "Exit-time upper bound (RFC3339 or RFC3339Nano, exclusive)"
// @Success 200 {object} response.EquitySeriesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /algotrade/equity [get]
func (h *Handler) GetEquitySeries(c echo.Context) error {
	if h.stateStore == nil || h.getEquitySeries == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "algotrade read API not configured",
			Status: "error",
		})
	}
	partition := strings.TrimSpace(c.QueryParam("partition"))
	runID := strings.TrimSpace(c.QueryParam("run_id"))
	if partition == "" && runID == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "partition or run_id is required",
			Status: "error",
		})
	}
	if partition == "" {
		if h.getRun == nil {
			return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
				Error:  "run read API not configured",
				Status: "error",
			})
		}
		run, ok, err := h.getRun.Execute(c.Request().Context(), dagruntimeusecase.GetRunRequest{RunID: runID})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:  err.Error(),
				Status: "error",
			})
		}
		if !ok {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:  "run not found",
				Status: "error",
			})
		}
		partition = run.Partition
	}
	fromTime, err := parseOptionalRFC3339(c.QueryParam("from"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "from must be RFC3339 or RFC3339Nano",
			Status: "error",
		})
	}
	toTime, err := parseOptionalRFC3339(c.QueryParam("to"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "to must be RFC3339 or RFC3339Nano",
			Status: "error",
		})
	}

	txn := h.stateStore.BeginTxn(domainstate.Partition(partition))
	defer txn.Rollback()
	closedTrades, ok := domainstate.Get(txn, algotrade.StateClosedTrades)
	if !ok {
		closedTrades = algotrade.ClosedTradesState{}
	}
	series, err := h.getEquitySeries.Execute(c.Request().Context(), dagruntimeusecase.GetEquitySeriesRequest{
		ClosedTrades: closedTrades,
		From:         fromTime,
		To:           toTime,
		Symbol:       strings.TrimSpace(c.QueryParam("symbol")),
		Side:         strings.TrimSpace(c.QueryParam("side")),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}

	resp := response.EquitySeriesResponse{
		Partition:   partition,
		RunID:       runID,
		MaxDrawdown: series.MaxDrawdown,
		TotalNetPnL: series.TotalNetPnL,
		Points:      make([]response.EquityPoint, 0, len(series.Points)),
	}
	for _, point := range series.Points {
		resp.Points = append(resp.Points, response.EquityPoint{
			Time:     point.Time,
			TradeID:  point.TradeID,
			Equity:   point.Equity,
			Drawdown: point.Drawdown,
		})
	}
	return c.JSON(http.StatusOK, resp)
}
