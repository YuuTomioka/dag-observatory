package usecase

import (
	"context"
	"testing"
	"time"

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

func TestListTradeResultsExecuteAppliesFiltersAndPaging(t *testing.T) {
	uc := &ListTradeResults{}

	from := time.Date(2026, 4, 5, 0, 30, 0, 0, time.UTC)
	to := time.Date(2026, 4, 5, 3, 0, 0, 0, time.UTC)
	items, err := uc.Execute(context.Background(), ListTradeResultsRequest{
		ClosedTrades: algotrade.ClosedTradesState{
			Items: []algotrade.ClosedTrade{
				{
					TradeID:  "trade:1",
					Symbol:   "USDJPY",
					Side:     algotrade.PositionSideLong,
					ExitTime: marketdata.MustParseUTCTime("2026-04-05T01:00:00Z"),
				},
				{
					TradeID:  "trade:2",
					Symbol:   "USDJPY",
					Side:     algotrade.PositionSideShort,
					ExitTime: marketdata.MustParseUTCTime("2026-04-05T02:00:00Z"),
				},
				{
					TradeID:  "trade:3",
					Symbol:   "EURUSD",
					Side:     algotrade.PositionSideLong,
					ExitTime: marketdata.MustParseUTCTime("2026-04-05T02:30:00Z"),
				},
			},
		},
		From:   &from,
		To:     &to,
		Symbol: "USDJPY",
		Side:   "short",
		Limit:  1,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("execute list trade results: %v", err)
	}
	if len(items) != 1 || items[0].TradeID != "trade:2" {
		t.Fatalf("expected one filtered item trade:2, got %#v", items)
	}
}

func TestGetEquitySeriesExecuteBuildsCumulativeEquityAndDrawdown(t *testing.T) {
	uc := &GetEquitySeries{}
	series, err := uc.Execute(context.Background(), GetEquitySeriesRequest{
		ClosedTrades: algotrade.ClosedTradesState{
			Items: []algotrade.ClosedTrade{
				{
					TradeID:  "trade:1",
					NetPnL:   10,
					ExitTime: marketdata.MustParseUTCTime("2026-04-05T01:00:00Z"),
				},
				{
					TradeID:  "trade:2",
					NetPnL:   -4,
					ExitTime: marketdata.MustParseUTCTime("2026-04-05T02:00:00Z"),
				},
				{
					TradeID:  "trade:3",
					NetPnL:   1,
					ExitTime: marketdata.MustParseUTCTime("2026-04-05T03:00:00Z"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("execute equity series: %v", err)
	}
	if len(series.Points) != 3 {
		t.Fatalf("expected 3 equity points, got %d", len(series.Points))
	}
	if series.Points[0].Equity != 10 || series.Points[1].Drawdown != 4 {
		t.Fatalf("unexpected series points: %#v", series.Points)
	}
	if series.MaxDrawdown != 4 || series.TotalNetPnL != 7 {
		t.Fatalf("unexpected summary: %#v", series)
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
