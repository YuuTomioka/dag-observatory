package handler

import (
	"net/http"

	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/algotrade"
	domainstate "dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/interface/http/algotrade/response"
	"dag-observatory/dag-core/internal/interface/http/dto"

	"github.com/labstack/echo/v4"
)

// GetSummary
// @Summary Get strategy summary
// @Description Returns strategy summary projection for the given runtime partition.
// @Tags algotrade
// @Produce json
// @Param partition query string true "Runtime partition"
// @Success 200 {object} response.GetSummaryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /algotrade/summary [get]
func (h *Handler) GetSummary(c echo.Context) error {
	if h.stateStore == nil || h.getSummary == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "algotrade summary API not configured",
			Status: "error",
		})
	}
	partition := c.QueryParam("partition")
	if partition == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "partition is required",
			Status: "error",
		})
	}

	txn := h.stateStore.BeginTxn(domainstate.Partition(partition))
	defer txn.Rollback()

	summaryState, ok := domainstate.Get(txn, algotrade.StateStrategySummary)
	if !ok {
		summaryState = algotrade.StrategySummaryState{}
	}
	summary, err := h.getSummary.Execute(c.Request().Context(), dagruntimeusecase.GetStrategySummaryRequest{
		Summary: summaryState,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}

	return c.JSON(http.StatusOK, response.GetSummaryResponse{
		Partition: partition,
		Summary: response.StrategySummary{
			StrategyID:      summary.StrategyID,
			WorkflowName:    summary.WorkflowName,
			WorkflowVersion: summary.WorkflowVersion,
			ParameterSetID:  summary.ParameterSetID,
			TradeCount:      summary.TradeCount,
			WinCount:        summary.WinCount,
			LossCount:       summary.LossCount,
			WinRate:         summary.WinRate,
			TotalNetPnL:     summary.TotalNetPnL,
			AverageWin:      summary.AverageWin,
			AverageLoss:     summary.AverageLoss,
			ProfitFactor:    summary.ProfitFactor,
			MaxDrawdown:     summary.MaxDrawdown,
			UpdatedAt:       summary.UpdatedAt,
		},
	})
}
