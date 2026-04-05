package usecase

import (
	"context"
	"slices"

	"dag-observatory/dag-core/internal/application/dagruntime/view"
	"dag-observatory/dag-core/internal/domain/algotrade"
)

type ListTradeResults struct{}

type ListTradeResultsRequest struct {
	ClosedTrades algotrade.ClosedTradesState
}

func (u *ListTradeResults) Execute(ctx context.Context, req ListTradeResultsRequest) ([]view.TradeView, error) {
	_ = ctx
	items := make([]view.TradeView, 0, len(req.ClosedTrades.Items))
	trades := slices.Clone(req.ClosedTrades.Items)
	slices.SortFunc(trades, func(a, b algotrade.ClosedTrade) int {
		switch {
		case a.ExitTime.After(b.ExitTime):
			return -1
		case b.ExitTime.After(a.ExitTime):
			return 1
		default:
			if a.TradeID < b.TradeID {
				return -1
			}
			if a.TradeID > b.TradeID {
				return 1
			}
			return 0
		}
	})
	for _, trade := range trades {
		items = append(items, view.TradeFromClosedTrade(trade))
	}
	return items, nil
}

type GetStrategySummary struct{}

type GetStrategySummaryRequest struct {
	Summary algotrade.StrategySummaryState
}

func (u *GetStrategySummary) Execute(ctx context.Context, req GetStrategySummaryRequest) (view.StrategySummaryView, error) {
	_ = ctx
	return view.StrategySummaryFromDomain(req.Summary.Summary), nil
}
