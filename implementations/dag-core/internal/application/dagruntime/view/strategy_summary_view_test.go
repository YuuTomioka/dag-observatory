package view

import (
	"testing"

	"dag-observatory/dag-core/internal/domain/algotrade"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

func TestStrategySummaryFromDomain(t *testing.T) {
	summary := algotrade.StrategySummary{
		StrategyID:      "breakout",
		WorkflowName:    "wf",
		WorkflowVersion: "v1",
		ParameterSetID:  "p1",
		TradeCount:      3,
		WinCount:        2,
		LossCount:       1,
		WinRate:         2.0 / 3.0,
		TotalNetPnL:     12,
		UpdatedAt:       marketdata.MustParseUTCTime("2026-04-05T03:00:00Z"),
	}

	view := StrategySummaryFromDomain(summary)
	if view.StrategyID != "breakout" || view.WorkflowVersion != "v1" {
		t.Fatalf("unexpected summary identity: %#v", view)
	}
	if view.TradeCount != 3 || view.TotalNetPnL != 12 {
		t.Fatalf("unexpected summary metrics: %#v", view)
	}
	if view.UpdatedAt == "" {
		t.Fatalf("expected updated_at formatting, got %#v", view)
	}
}
