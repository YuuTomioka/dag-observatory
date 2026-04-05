package usecase

import (
	"context"
	"testing"

	"dag-observatory/dag-core/internal/domain/algotrade"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

func TestListTradeResultsExecuteReturnsNewestFirst(t *testing.T) {
	uc := &ListTradeResults{}

	items, err := uc.Execute(context.Background(), ListTradeResultsRequest{
		ClosedTrades: algotrade.ClosedTradesState{
			Items: []algotrade.ClosedTrade{
				{
					TradeID:  "trade:1",
					Symbol:   "USDJPY",
					ExitTime: marketdata.MustParseUTCTime("2026-04-05T01:00:00Z"),
				},
				{
					TradeID:  "trade:2",
					Symbol:   "EURUSD",
					ExitTime: marketdata.MustParseUTCTime("2026-04-05T02:00:00Z"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("execute list trade results: %v", err)
	}
	if len(items) != 2 || items[0].TradeID != "trade:2" {
		t.Fatalf("expected newest trade first, got %#v", items)
	}
}

func TestGetStrategySummaryExecuteProjectsSummaryView(t *testing.T) {
	uc := &GetStrategySummary{}

	summary, err := uc.Execute(context.Background(), GetStrategySummaryRequest{
		Summary: algotrade.StrategySummaryState{
			Summary: algotrade.StrategySummary{
				StrategyID:      "breakout",
				WorkflowName:    "wf",
				WorkflowVersion: "v1",
				TradeCount:      3,
				TotalNetPnL:     12,
			},
		},
	})
	if err != nil {
		t.Fatalf("execute get strategy summary: %v", err)
	}
	if summary.StrategyID != "breakout" || summary.TradeCount != 3 || summary.TotalNetPnL != 12 {
		t.Fatalf("unexpected summary view: %#v", summary)
	}
}
