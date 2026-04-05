package factory

import (
	"context"
	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/algotrade"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/engine"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"
	"dag-observatory/dag-core/internal/domain/dagruntime/policy"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/marketdata"
	artifactinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
	"testing"
	"time"
)

func TestPositionBreakevenAndSignalExitBasicIntegration(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	breakevenNode, err := registry.Build(spec.NodeSpec{
		ID:   "breakeven",
		Kind: "position_breakeven",
		Config: map[string]any{
			"position_node_id": "position_load",
		},
	})
	if err != nil {
		t.Fatalf("build position_breakeven: %v", err)
	}
	exitNode, err := registry.Build(spec.NodeSpec{
		ID:   "exit",
		Kind: "signal_exit_basic",
		Config: map[string]any{
			"position_node_id":  "position_load",
			"stop_loss_node_id": "breakeven",
		},
	})
	if err != nil {
		t.Fatalf("build signal_exit_basic: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, positionSnapshotOutputKey("position_load"), algotrade.PositionSnapshot{
		HasPosition: true,
		Side:        algotrade.PositionSideLong,
		Size:        1.0,
		EntryPrice:  marketdata.NewPriceFromRaw(1000),
	})
	artifact.Set(writer, usecase.InputKeyMarketOHLCVBars, []marketdata.OHLCV{
		{
			Open:  marketdata.NewPriceFromRaw(1000),
			High:  marketdata.NewPriceFromRaw(1001),
			Low:   marketdata.NewPriceFromRaw(995),
			Close: marketdata.NewPriceFromRaw(998),
		},
	})

	mem := stateinfra.NewMemoryStore()
	txn := mem.BeginTxn("test")
	state.StageWrite(txn, algotrade.StateOpenPositions, algotrade.OpenPositionsState{
		Items: []algotrade.OpenPosition{
			{
				PositionID: "p-1",
				Symbol:     "USDJPY",
				Snapshot:   artifact.MustGet(artifacts.View(), positionSnapshotOutputKey("position_load")),
			},
		},
	})

	if err := breakevenNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run position_breakeven: %v", err)
	}
	if err := exitNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run signal_exit_basic: %v", err)
	}
	decision := artifact.MustGet(artifacts.View(), signalExitDecisionOutputKey("exit"))
	if !decision.ShouldExit {
		t.Fatalf("expected should_exit=true via breakeven stop loss, got %#v", decision)
	}
}

func TestPositionBreakevenFactoryBuildRequiresPositionNodeID(t *testing.T) {
	t.Parallel()

	factory := &PositionBreakevenFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "breakeven",
		Kind:   "position_breakeven",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing position_node_id")
	}
}

func TestPositionSnapshotLoadFactoryBuildRejectsUnknownConfig(t *testing.T) {
	t.Parallel()

	factory := &PositionSnapshotLoadFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:   "position_load",
		Kind: "position_snapshot_load",
		Config: map[string]any{
			"unexpected": true,
		},
	})
	if err == nil {
		t.Fatal("expected validation error for unknown config key")
	}
}

func TestPositionSnapshotLoadRunReadsOpenPositionState(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	loadNode, err := registry.Build(spec.NodeSpec{
		ID:   "position_load",
		Kind: "position_snapshot_load",
	})
	if err != nil {
		t.Fatalf("build position_snapshot_load: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	mem := stateinfra.NewMemoryStore()
	txn := mem.BeginTxn("test")
	state.StageWrite(txn, algotrade.StateOpenPositions, algotrade.OpenPositionsState{
		Items: []algotrade.OpenPosition{
			{
				PositionID: "p-1",
				Symbol:     "USDJPY",
				Snapshot: algotrade.PositionSnapshot{
					HasPosition: true,
					Side:        algotrade.PositionSideLong,
					Size:        0.6,
					EntryPrice:  marketdata.NewPriceFromRaw(1111),
				},
			},
		},
	})

	if err := loadNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run position_snapshot_load: %v", err)
	}

	snapshot := artifact.MustGet(artifacts.View(), positionSnapshotOutputKey("position_load"))
	if !snapshot.HasPosition || snapshot.EntryPrice.Raw() != 1111 {
		t.Fatalf("expected snapshot from open_positions state, got %#v", snapshot)
	}
}

func TestPositionCloseToClosedTradeRunProducesClosedTrade(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	closeNode, err := registry.Build(spec.NodeSpec{
		ID:   "close_trade",
		Kind: "position_close_to_closed_trade",
		Config: map[string]any{
			"position_node_id": "position_load",
			"exit_node_id":     "exit",
		},
	})
	if err != nil {
		t.Fatalf("build position_close_to_closed_trade: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, positionSnapshotOutputKey("position_load"), algotrade.PositionSnapshot{
		HasPosition: true,
		Side:        algotrade.PositionSideLong,
		Size:        1.0,
		EntryPrice:  marketdata.NewPriceFromRaw(1000),
	})
	artifact.Set(writer, signalExitDecisionOutputKey("exit"), algotrade.ExitDecision{
		ShouldExit: true,
		Reason:     "stop_loss_hit",
	})
	artifact.Set(writer, usecase.InputKeyMarketOHLCVBars, []marketdata.OHLCV{
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-05T00:59:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-05T01:00:00Z"),
			Open:      marketdata.NewPriceFromRaw(995),
			High:      marketdata.NewPriceFromRaw(1000),
			Low:       marketdata.NewPriceFromRaw(979),
			Close:     marketdata.NewPriceFromRaw(980),
		},
	})

	mem := stateinfra.NewMemoryStore()
	txn := mem.BeginTxn("test")
	state.StageWrite(txn, algotrade.StateOpenPositions, algotrade.OpenPositionsState{
		Items: []algotrade.OpenPosition{
			{
				PositionID:      "pos:exec:intent:1",
				IntentID:        "intent:1",
				ExecutionID:     "exec:intent:1:1",
				Symbol:          "USDJPY",
				Side:            algotrade.PositionSideLong,
				Size:            1.0,
				EntryPrice:      marketdata.NewPriceFromRaw(1000),
				EntryTime:       marketdata.MustParseUTCTime("2026-04-05T00:00:00Z"),
				EntryReason:     "entry_allowed",
				StrategyID:      "breakout",
				WorkflowName:    "dagruntime.breakout_long_v1_extended",
				WorkflowVersion: "v1",
				ParameterSetID:  "p1",
				Snapshot: algotrade.PositionSnapshot{
					HasPosition: true,
					Side:        algotrade.PositionSideLong,
					Size:        1.0,
					EntryPrice:  marketdata.NewPriceFromRaw(1000),
				},
			},
		},
	})

	if err := closeNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run position_close_to_closed_trade: %v", err)
	}

	trade := artifact.MustGet(artifacts.View(), closedTradeOutputKey("close_trade"))
	if trade.TradeID != "trade:intent:1:1" || trade.IntentID != "intent:1" {
		t.Fatalf("expected closed trade identity, got %#v", trade)
	}
	if trade.ExitReason != "stop_loss_hit" || trade.ExitPrice.Raw() != 980 {
		t.Fatalf("expected exit fields to be set, got %#v", trade)
	}
	if trade.NetPnL != -20 {
		t.Fatalf("expected net pnl -20, got %#v", trade)
	}
	if trade.WorkflowName != "dagruntime.breakout_long_v1_extended" || trade.ParameterSetID != "p1" {
		t.Fatalf("expected workflow context propagation, got %#v", trade)
	}
}

func TestClosedTradeStoreRunAppendsStateOnce(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	storeNode, err := registry.Build(spec.NodeSpec{
		ID:   "closed_store",
		Kind: "closed_trade_store",
		Config: map[string]any{
			"trade_node_id": "close_trade",
		},
	})
	if err != nil {
		t.Fatalf("build closed_trade_store: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, closedTradeOutputKey("close_trade"), algotrade.ClosedTrade{
		TradeID:         "trade:intent:1:1",
		IntentID:        "intent:1",
		PositionID:      "pos:exec:intent:1",
		Symbol:          "USDJPY",
		Side:            algotrade.PositionSideLong,
		Size:            1.0,
		EntryTime:       marketdata.MustParseUTCTime("2026-04-05T00:00:00Z"),
		ExitTime:        marketdata.MustParseUTCTime("2026-04-05T01:00:00Z"),
		EntryPrice:      marketdata.NewPriceFromRaw(1000),
		ExitPrice:       marketdata.NewPriceFromRaw(980),
		NetPnL:          -20,
		ExitReason:      "stop_loss_hit",
		HoldingDuration: "1h0m0s",
	})

	mem := stateinfra.NewMemoryStore()
	txn := mem.BeginTxn("test")
	if err := storeNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("first run closed_trade_store: %v", err)
	}
	if err := storeNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("second run closed_trade_store: %v", err)
	}

	closedTrades := state.MustGet(txn, algotrade.StateClosedTrades)
	if len(closedTrades.Items) != 1 {
		t.Fatalf("expected one closed trade stored, got %#v", closedTrades)
	}
	if closedTrades.Items[0].TradeID != "trade:intent:1:1" {
		t.Fatalf("expected stored trade identity, got %#v", closedTrades.Items[0])
	}
}

func TestDailyPnLUpdateRunAccumulatesNetPnLByExitDay(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	updateNode, err := registry.Build(spec.NodeSpec{
		ID:   "daily_pnl",
		Kind: "daily_pnl_update",
		Config: map[string]any{
			"trade_node_id": "close_trade",
		},
	})
	if err != nil {
		t.Fatalf("build daily_pnl_update: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, closedTradeOutputKey("close_trade"), algotrade.ClosedTrade{
		TradeID:   "trade:intent:1:1",
		IntentID:  "intent:1",
		ExitTime:  marketdata.MustParseUTCTime("2026-04-05T01:00:00Z"),
		GrossPnL:  -20,
		NetPnL:    -20,
		ExitPrice: marketdata.NewPriceFromRaw(980),
	})

	mem := stateinfra.NewMemoryStore()
	txn := mem.BeginTxn("test")
	if err := updateNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("first run daily_pnl_update: %v", err)
	}

	daily := state.MustGet(txn, algotrade.StateDailyPnL)
	if daily.TradingDay != "2026-04-05" || daily.RealizedPnL != -20 {
		t.Fatalf("expected realized pnl update, got %#v", daily)
	}
}

func TestOpenPositionCloseRunRemovesClosedPosition(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	closeNode, err := registry.Build(spec.NodeSpec{
		ID:   "position_close",
		Kind: "open_position_close",
		Config: map[string]any{
			"trade_node_id": "close_trade",
		},
	})
	if err != nil {
		t.Fatalf("build open_position_close: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, closedTradeOutputKey("close_trade"), algotrade.ClosedTrade{
		TradeID:    "trade:intent:1:1",
		IntentID:   "intent:1",
		PositionID: "pos:exec:intent:1",
	})

	mem := stateinfra.NewMemoryStore()
	txn := mem.BeginTxn("test")
	state.StageWrite(txn, algotrade.StateOpenPositions, algotrade.OpenPositionsState{
		Items: []algotrade.OpenPosition{
			{PositionID: "pos:exec:intent:1", Symbol: "USDJPY"},
			{PositionID: "pos:exec:intent:2", Symbol: "EURUSD"},
		},
	})

	if err := closeNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run open_position_close: %v", err)
	}

	openPositions := state.MustGet(txn, algotrade.StateOpenPositions)
	if len(openPositions.Items) != 1 || openPositions.Items[0].PositionID != "pos:exec:intent:2" {
		t.Fatalf("expected closed position removal, got %#v", openPositions)
	}
}

func TestPositionTimeoutExitFactoryBuildRequiresPositionNodeID(t *testing.T) {
	t.Parallel()

	factory := &PositionTimeoutExitFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "timeout",
		Kind:   "position_timeout_exit",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing position_node_id")
	}
}

func TestPositionTimeoutExitRunWithoutPosition(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	timeoutNode, err := registry.Build(spec.NodeSpec{
		ID:   "timeout",
		Kind: "position_timeout_exit",
		Config: map[string]any{
			"position_node_id": "position_load",
		},
	})
	if err != nil {
		t.Fatalf("build position_timeout_exit: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, positionSnapshotOutputKey("position_load"), algotrade.PositionSnapshot{
		HasPosition: false,
		Side:        algotrade.PositionSideFlat,
		Size:        0,
		EntryPrice:  marketdata.Price(0),
	})

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	if err := timeoutNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run position_timeout_exit: %v", err)
	}
	got := artifact.MustGet(artifacts.View(), signalExitDecisionOutputKey("timeout"))
	if got.ShouldExit {
		t.Fatalf("expected should_exit=false, got %#v", got)
	}
	if got.Reason != "no_position" {
		t.Fatalf("expected reason=no_position, got %q", got.Reason)
	}
}

func TestPositionTrailingStopFactoryBuildRequiresPositionNodeID(t *testing.T) {
	t.Parallel()

	factory := &PositionTrailingStopFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "trailing",
		Kind:   "position_trailing_stop",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing position_node_id")
	}
}

func TestPositionTrailingStopRunProducesTighterDistance(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	trailingNode, err := registry.Build(spec.NodeSpec{
		ID:   "trailing",
		Kind: "position_trailing_stop",
		Config: map[string]any{
			"position_node_id": "position_load",
		},
	})
	if err != nil {
		t.Fatalf("build position_trailing_stop: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, positionSnapshotOutputKey("position_load"), algotrade.PositionSnapshot{
		HasPosition: true,
		Side:        algotrade.PositionSideLong,
		Size:        1.0,
		EntryPrice:  marketdata.NewPriceFromRaw(1000),
	})
	artifact.Set(writer, usecase.InputKeyMarketOHLCVBars, []marketdata.OHLCV{
		{
			Open:  marketdata.NewPriceFromRaw(1000),
			High:  marketdata.NewPriceFromRaw(1030),
			Low:   marketdata.NewPriceFromRaw(999),
			Close: marketdata.NewPriceFromRaw(1020),
		},
	})
	mem := stateinfra.NewMemoryStore()
	txn := mem.BeginTxn("test")
	state.StageWrite(txn, algotrade.StateOpenPositions, algotrade.OpenPositionsState{
		Items: []algotrade.OpenPosition{
			{
				PositionID: "p-1",
				Symbol:     "USDJPY",
				Snapshot:   artifact.MustGet(artifacts.View(), positionSnapshotOutputKey("position_load")),
			},
		},
	})

	if err := trailingNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run position_trailing_stop: %v", err)
	}
	stopDistance := artifact.MustGet(artifacts.View(), riskStopLossDistanceOutputKey("trailing"))
	if stopDistance.Raw() <= 1 {
		t.Fatalf("expected trailing stop distance > 1, got %d", stopDistance.Raw())
	}
}

func TestStateNodePartitionIsolation(t *testing.T) {
	t.Parallel()

	positionInputKey := artifact.Key[algotrade.PositionSnapshot]{
		Name:     "test.position_input",
		StableID: "artifact:test.position_input.v1",
	}
	positionStateKey := state.Key[algotrade.PositionSnapshot]{
		Name:     "position.snapshot",
		StableID: "state:dagruntime.position.snapshot.v1",
	}
	n := &writePositionStateNode{
		inputKey: positionInputKey,
		stateKey: positionStateKey,
	}
	compiled := pipeline.Compiled{
		Name:  "test.partition_isolation",
		Order: []node.Node{n},
		Nodes: []node.Node{n},
	}

	artifactStore := artifactinfra.NewMemoryStore()
	stateStore := stateinfra.NewMemoryStore()
	runner := &engine.Runner{
		ArtifactStore: artifactStore,
		StateStore:    stateStore,
		Policy:        policy.Policy{},
	}

	run := func(partition state.Partition, size float64, entryRaw int64) {
		t.Helper()
		inputs := engine.InputMap{
			positionInputKey: algotrade.PositionSnapshot{
				HasPosition: true,
				Side:        algotrade.PositionSideLong,
				Size:        size,
				EntryPrice:  marketdata.NewPriceFromRaw(entryRaw),
			},
		}
		event := events.Event{
			EventID:   "partition-isolation-" + string(partition),
			EventTime: time.Now().UTC(),
			Partition: partition,
			Type:      "task.requested",
		}
		if err := runner.RunCycle(context.Background(), compiled, inputs, partition, event); err != nil {
			t.Fatalf("run cycle for partition %s: %v", partition, err)
		}
	}

	partA := state.Partition("strategy-a|USDJPY")
	partB := state.Partition("strategy-a|EURUSD")
	run(partA, 1.25, 1011)
	run(partB, 0.75, 1502)

	txnA := stateStore.BeginTxn(partA)
	gotA, ok := state.Get(txnA, positionStateKey)
	if !ok {
		t.Fatal("expected state in partition A")
	}
	if gotA.Size != 1.25 || gotA.EntryPrice.Raw() != 1011 {
		t.Fatalf("unexpected partition A state: %#v", gotA)
	}
	txnA.Rollback()

	txnB := stateStore.BeginTxn(partB)
	gotB, ok := state.Get(txnB, positionStateKey)
	if !ok {
		t.Fatal("expected state in partition B")
	}
	if gotB.Size != 0.75 || gotB.EntryPrice.Raw() != 1502 {
		t.Fatalf("unexpected partition B state: %#v", gotB)
	}
	txnB.Rollback()
}
