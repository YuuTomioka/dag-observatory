package algotrade

import (
	"testing"

	"dag-observatory/dag-core/internal/domain/marketdata"
)

func TestTradeIntentCarriesWorkflowContext(t *testing.T) {
	t.Parallel()

	intent := TradeIntent{
		IntentID:        "intent:breakout:USDJPY:v1:2026-04-05T00:00:00Z:1",
		Symbol:          "USDJPY",
		Action:          OrderActionBuy,
		Reason:          "entry_allowed",
		PositionSize:    1.25,
		InitialStopLoss: marketdata.NewPriceFromRaw(25),
		WorkflowName:    "dagruntime.breakout_long_v1_extended",
		WorkflowVersion: "v1",
		CreatedAt:       marketdata.MustParseUTCTime("2026-04-05T00:00:00Z"),
	}

	if intent.IntentID == "" {
		t.Fatal("expected non-empty intent id")
	}
	if intent.Action != OrderActionBuy {
		t.Fatalf("expected buy action, got %q", intent.Action)
	}
	if intent.WorkflowName == "" || intent.WorkflowVersion == "" {
		t.Fatalf("expected workflow context, got %#v", intent)
	}
	if intent.CreatedAt.IsZero() {
		t.Fatalf("expected created_at to be set, got %#v", intent)
	}
}

func TestExecutionStatusConstants(t *testing.T) {
	t.Parallel()

	got := []ExecutionStatus{
		ExecutionStatusRequested,
		ExecutionStatusSubmitted,
		ExecutionStatusAccepted,
		ExecutionStatusPartiallyFilled,
		ExecutionStatusFilled,
		ExecutionStatusRejected,
		ExecutionStatusCancelled,
	}
	want := []string{
		"requested",
		"submitted",
		"accepted",
		"partially_filled",
		"filled",
		"rejected",
		"cancelled",
	}
	for i := range want {
		if string(got[i]) != want[i] {
			t.Fatalf("unexpected execution status at %d: got=%q want=%q", i, got[i], want[i])
		}
	}
}

func TestOrderRequestToTradeIntent(t *testing.T) {
	t.Parallel()

	req := OrderRequest{
		Action:           OrderActionBuy,
		Reason:           "entry_allowed",
		PositionSize:     0.8,
		StopLossDistance: marketdata.NewPriceFromRaw(15),
	}
	at := marketdata.MustParseUTCTime("2026-04-05T00:00:00Z")

	intent := req.ToTradeIntent(
		"intent:breakout:USDJPY:v1:2026-04-05T00:00:00Z:1",
		"USDJPY",
		"dagruntime.breakout_long_v1_extended",
		"v1",
		at,
	)

	if intent.IntentID == "" || intent.Symbol != "USDJPY" {
		t.Fatalf("unexpected trade intent identity: %#v", intent)
	}
	if intent.InitialStopLoss.Raw() != 15 {
		t.Fatalf("expected initial stop loss to be propagated, got %#v", intent)
	}
	if !intent.CreatedAt.Equal(at) {
		t.Fatalf("expected created_at propagation, got %#v", intent)
	}
}

func TestMarketOrderRequestToExecutionResult(t *testing.T) {
	t.Parallel()

	req := MarketOrderRequest{
		Submitted: true,
		OrderID:   "market_exec-1",
		Action:    OrderActionBuy,
		Size:      1.2,
		Reason:    "submitted",
		SourceID:  "buy:entry_allowed:1.2:15",
	}
	at := marketdata.MustParseUTCTime("2026-04-05T00:00:00Z")

	result := req.ToExecutionResult("exec:intent:1:1", "intent:1", at)
	if result.Status != ExecutionStatusSubmitted {
		t.Fatalf("expected submitted status, got %#v", result)
	}
	if result.OrderID != "market_exec-1" || result.SourceID == "" {
		t.Fatalf("expected execution identity propagation, got %#v", result)
	}
}

func TestFillResultToExecutionResult(t *testing.T) {
	t.Parallel()

	fill := FillResult{
		Filled:  true,
		OrderID: "market_exec-1",
		Action:  OrderActionBuy,
		Size:    1.2,
		Reason:  "filled",
	}
	requestedAt := marketdata.MustParseUTCTime("2026-04-05T00:00:00Z")
	updatedAt := marketdata.MustParseUTCTime("2026-04-05T00:01:00Z")

	result := fill.ToExecutionResult("exec:intent:1:1", "intent:1", requestedAt, updatedAt)
	if result.Status != ExecutionStatusFilled {
		t.Fatalf("expected filled status, got %#v", result)
	}
	if result.FilledSize != 1.2 {
		t.Fatalf("expected filled size propagation, got %#v", result)
	}
	if !result.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected updated_at propagation, got %#v", result)
	}
}

func TestClosedTradeCapturesValidationFacts(t *testing.T) {
	t.Parallel()

	trade := ClosedTrade{
		TradeID:         "trade:intent:1:1",
		IntentID:        "intent:1",
		PositionID:      "pos:exec:1",
		Symbol:          "USDJPY",
		Side:            PositionSideLong,
		Size:            1.0,
		EntryTime:       marketdata.MustParseUTCTime("2026-04-05T00:00:00Z"),
		ExitTime:        marketdata.MustParseUTCTime("2026-04-05T01:00:00Z"),
		EntryPrice:      marketdata.NewPriceFromRaw(1000),
		ExitPrice:       marketdata.NewPriceFromRaw(1020),
		NetPnL:          20.0,
		ExitReason:      "take_profit",
		WorkflowName:    "dagruntime.breakout_long_v1_extended",
		WorkflowVersion: "v1",
	}

	if trade.TradeID == "" || trade.IntentID == "" {
		t.Fatalf("expected ids to be set, got %#v", trade)
	}
	if trade.Side != PositionSideLong {
		t.Fatalf("expected long side, got %q", trade.Side)
	}
	if !trade.ExitTime.After(trade.EntryTime) {
		t.Fatalf("expected exit after entry, got %#v", trade)
	}
	if trade.NetPnL <= 0 {
		t.Fatalf("expected positive pnl, got %#v", trade)
	}
	if trade.WorkflowName == "" || trade.WorkflowVersion == "" {
		t.Fatalf("expected workflow context, got %#v", trade)
	}
}
