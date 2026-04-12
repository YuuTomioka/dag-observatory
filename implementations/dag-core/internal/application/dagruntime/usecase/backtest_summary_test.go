package usecase

import (
	"context"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

func TestGetRunBacktestSummaryExecute(t *testing.T) {
	now := time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)
	reader := &fakeRunReader{
		runs: []port.RunRecord{
			{
				RunID:     "run-1",
				Partition: state.Partition("p1"),
				EventTime: now,
				StartedAt: now,
				EndedAt:   now.Add(time.Second),
				Duration:  time.Second,
			},
		},
		steps: map[string][]events.NodeExecutionEvent{
			"run-1": {
				{
					SequenceNo: 1,
					StateDiff: []events.StateDiffField{
						{Field: "strategy_summary.trade_count", After: 2.0},
						{Field: "strategy_summary.win_rate", After: 0.5},
						{Field: "strategy_summary.total_net_pnl", After: 8.2},
						{Field: "strategy_summary.max_drawdown", After: 1.7},
						{Field: "equity_series.count", After: 10.0},
					},
				},
			},
		},
	}

	uc := &GetRunBacktestSummary{Reader: reader}
	summary, ok, err := uc.Execute(context.Background(), GetRunBacktestSummaryRequest{RunID: "run-1"})
	if err != nil {
		t.Fatalf("get run backtest summary failed: %v", err)
	}
	if !ok {
		t.Fatal("expected summary to exist")
	}
	if summary.TradeCount != 2 || summary.TotalNetPnL != 8.2 || summary.MaxDrawdown != 1.7 {
		t.Fatalf("unexpected summary values: %#v", summary)
	}
	if summary.EquityPointCount != 10 {
		t.Fatalf("expected equity_point_count=10, got %d", summary.EquityPointCount)
	}
}
