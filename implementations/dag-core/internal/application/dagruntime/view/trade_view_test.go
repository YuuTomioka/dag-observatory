package view

import (
	"testing"

	"dag-observatory/dag-core/internal/domain/algotrade"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

func TestTradeFromClosedTrade(t *testing.T) {
	trade := algotrade.ClosedTrade{
		TradeID:         "trade:intent:1:1",
		IntentID:        "intent:1",
		PositionID:      "pos:1",
		Symbol:          "USDJPY",
		Side:            algotrade.PositionSideLong,
		Size:            1.2,
		EntryTime:       marketdata.MustParseUTCTime("2026-04-05T00:00:00Z"),
		ExitTime:        marketdata.MustParseUTCTime("2026-04-05T01:00:00Z"),
		EntryPrice:      marketdata.NewPriceFromRaw(1000),
		ExitPrice:       marketdata.NewPriceFromRaw(1015),
		NetPnL:          18,
		ExitReason:      "take_profit",
		WorkflowName:    "wf",
		WorkflowVersion: "v1",
		ParameterSetID:  "p1",
	}

	view := TradeFromClosedTrade(trade)
	if view.TradeID != "trade:intent:1:1" || view.Symbol != "USDJPY" {
		t.Fatalf("unexpected trade view identity: %#v", view)
	}
	if view.EntryPriceRaw != 1000 || view.ExitPriceRaw != 1015 || view.NetPnL != 18 {
		t.Fatalf("unexpected trade economics: %#v", view)
	}
	if view.WorkflowVersion != "v1" || view.ParameterSetID != "p1" {
		t.Fatalf("expected workflow context projection: %#v", view)
	}
}
