package handler

import (
	"net/http"

	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/application/dagruntime/view"
	"dag-observatory/dag-core/internal/domain/algotrade"
	domainstate "dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/interface/http/algotrade/response"
	"dag-observatory/dag-core/internal/interface/http/dto"

	"github.com/labstack/echo/v4"
)

// ListTrades
// @Summary List closed trades
// @Description Returns closed-trade facts for the given runtime partition. This is a snapshot read path for result confirmation, not an observability event feed.
// @Tags algotrade
// @Produce json
// @Param partition query string true "Runtime partition"
// @Success 200 {object} response.ListTradesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /algotrade/trades [get]
func (h *Handler) ListTrades(c echo.Context) error {
	if h.stateStore == nil || h.listTradeResults == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "algotrade read API not configured",
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

	closedTrades, ok := domainstate.Get(txn, algotrade.StateClosedTrades)
	if !ok {
		closedTrades = algotrade.ClosedTradesState{}
	}

	items, err := h.listTradeResults.Execute(c.Request().Context(), dagruntimeusecase.ListTradeResultsRequest{
		ClosedTrades: closedTrades,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}

	resp := response.ListTradesResponse{
		Partition: partition,
		Items:     make([]response.TradeItem, 0, len(items)),
	}
	for _, item := range items {
		resp.Items = append(resp.Items, mapTradeItem(item))
	}
	return c.JSON(http.StatusOK, resp)
}

func mapTradeItem(item view.TradeView) response.TradeItem {
	return response.TradeItem{
		TradeID:         item.TradeID,
		IntentID:        item.IntentID,
		PositionID:      item.PositionID,
		Symbol:          item.Symbol,
		Side:            item.Side,
		Size:            item.Size,
		EntryTime:       item.EntryTime,
		ExitTime:        item.ExitTime,
		EntryPriceRaw:   item.EntryPriceRaw,
		ExitPriceRaw:    item.ExitPriceRaw,
		NetPnL:          item.NetPnL,
		ExitReason:      item.ExitReason,
		WorkflowName:    item.WorkflowName,
		WorkflowVersion: item.WorkflowVersion,
		ParameterSetID:  item.ParameterSetID,
	}
}
