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
