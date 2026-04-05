package algotrade

import (
	"testing"

	domainstate "dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/marketdata"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
)

func TestStateKeysStableIDs(t *testing.T) {
	t.Parallel()

	if got := StateOpenPositions.Raw().StableID; got != "state:algotrade.open_positions.v1" {
		t.Fatalf("unexpected open positions stable id: %s", got)
	}
	if got := StatePendingOrders.Raw().StableID; got != "state:algotrade.pending_orders.v1" {
		t.Fatalf("unexpected pending orders stable id: %s", got)
	}
	if got := StateDailyPnL.Raw().StableID; got != "state:algotrade.daily_pnl.v1" {
		t.Fatalf("unexpected daily pnl stable id: %s", got)
	}
	if got := StateClosedTrades.Raw().StableID; got != "state:algotrade.closed_trades.v1" {
		t.Fatalf("unexpected closed trades stable id: %s", got)
	}
	if got := StateStrategySummary.Raw().StableID; got != "state:algotrade.strategy_summary.v1" {
		t.Fatalf("unexpected strategy summary stable id: %s", got)
	}
	if got := StateLossStreak.Raw().StableID; got != "state:algotrade.loss_streak.v1" {
		t.Fatalf("unexpected loss streak stable id: %s", got)
	}
}

func TestClosedTradesStateRoundTrip(t *testing.T) {
	t.Parallel()

	store := stateinfra.NewMemoryStore()
	txn := store.BeginTxn(domainstate.Partition("strategy:breakout:USDJPY"))
	want := ClosedTradesState{
		Items: []ClosedTrade{
			{
				TradeID:         "trade:intent:1:1",
				IntentID:        "intent:1",
				Symbol:          "USDJPY",
				Side:            PositionSideLong,
				Size:            1.0,
				ExitReason:      "take_profit",
				WorkflowName:    "dagruntime.breakout_long_v1_extended",
				WorkflowVersion: "v1",
			},
		},
	}
	domainstate.StageWrite(txn, StateClosedTrades, want)

	got, ok := domainstate.Get(txn, StateClosedTrades)
	if !ok {
		t.Fatal("expected closed trades state to be staged")
	}
	if len(got.Items) != 1 || got.Items[0].TradeID != want.Items[0].TradeID {
		t.Fatalf("unexpected closed trades state: %#v", got)
	}
}

func TestStrategySummaryStateRoundTrip(t *testing.T) {
	t.Parallel()

	store := stateinfra.NewMemoryStore()
	txn := store.BeginTxn(domainstate.Partition("strategy:breakout:USDJPY"))
	want := StrategySummaryState{
		Summary: StrategySummary{
			StrategyID:      "breakout",
			WorkflowName:    "dagruntime.breakout_long_v1_extended",
			WorkflowVersion: "v1",
			TradeCount:      3,
			WinCount:        2,
			LossCount:       1,
			WinRate:         2.0 / 3.0,
			TotalNetPnL:     42.0,
		},
	}
	domainstate.StageWrite(txn, StateStrategySummary, want)

	got, ok := domainstate.Get(txn, StateStrategySummary)
	if !ok {
		t.Fatal("expected strategy summary state to be staged")
	}
	if got.Summary.TradeCount != 3 || got.Summary.TotalNetPnL != 42.0 {
		t.Fatalf("unexpected strategy summary state: %#v", got)
	}
}

func TestDailyPnLStateToSnapshot(t *testing.T) {
	t.Parallel()

	pnl := DailyPnLState{
		TradingDay:    "2026-04-05",
		RealizedPnL:   10.5,
		UnrealizedPnL: 2.25,
		LossLimitHit:  true,
	}

	got := pnl.ToSnapshot()
	if got.TradingDay != "2026-04-05" || got.RealizedPnL != 10.5 || !got.LossLimitHit {
		t.Fatalf("unexpected pnl snapshot: %#v", got)
	}
}

func TestOpenPositionSnapshotViewFallsBackToCoreFields(t *testing.T) {
	t.Parallel()

	position := OpenPosition{
		PositionID:  "pos:exec:1",
		Symbol:      "USDJPY",
		Side:        PositionSideLong,
		Size:        0.7,
		EntryPrice:  marketdata.NewPriceFromRaw(1234),
		EntryReason: "entry_allowed",
	}

	got := position.SnapshotView()
	if !got.HasPosition || got.Side != PositionSideLong || got.Size != 0.7 || got.EntryPrice.Raw() != 1234 {
		t.Fatalf("unexpected fallback snapshot: %#v", got)
	}
}

func TestPendingOrderRoundTripKeepsExecutionLinkFields(t *testing.T) {
	t.Parallel()

	store := stateinfra.NewMemoryStore()
	txn := store.BeginTxn(domainstate.Partition("strategy:breakout:USDJPY"))
	want := PendingOrdersState{
		Items: []PendingOrder{
			{
				OrderID:     "order-1",
				IntentID:    "intent-1",
				ExecutionID: "exec:order-1:1",
				Symbol:      "USDJPY",
				Intent: TradeIntent{
					IntentID:     "intent-1",
					Symbol:       "USDJPY",
					Action:       OrderActionBuy,
					Reason:       "entry_allowed",
					PositionSize: 0.5,
				},
				Execution: ExecutionResult{
					ExecutionID:   "exec:order-1:1",
					IntentID:      "intent-1",
					OrderID:       "order-1",
					Status:        ExecutionStatusSubmitted,
					Action:        OrderActionBuy,
					RequestedSize: 0.5,
					Reason:        "submitted",
					RequestedAt:   marketdata.MustParseUTCTime("2026-04-05T00:00:00Z"),
					UpdatedAt:     marketdata.MustParseUTCTime("2026-04-05T00:00:00Z"),
				},
				SourceID: "buy:entry_allowed:0.5:0",
			},
		},
	}
	domainstate.StageWrite(txn, StatePendingOrders, want)

	got, ok := domainstate.Get(txn, StatePendingOrders)
	if !ok {
		t.Fatal("expected pending orders state to be staged")
	}
	if got.Items[0].IntentID != "intent-1" || got.Items[0].ExecutionID != "exec:order-1:1" {
		t.Fatalf("unexpected pending order linkage: %#v", got)
	}
	if got.Items[0].Intent.Reason != "entry_allowed" || got.Items[0].Execution.Status != ExecutionStatusSubmitted {
		t.Fatalf("unexpected pending order payloads: %#v", got)
	}
}
