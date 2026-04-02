package factory

import (
	"context"
	"testing"
	"time"

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
)

type writePositionStateNode struct {
	inputKey artifact.Key[algotrade.PositionSnapshot]
	stateKey state.Key[algotrade.PositionSnapshot]
}

func (n *writePositionStateNode) Name() string { return "write.position.state" }
func (n *writePositionStateNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{n.inputKey}
}
func (n *writePositionStateNode) Provides() []artifact.AnyKey { return nil }
func (n *writePositionStateNode) Reads() []state.AnyKey       { return nil }
func (n *writePositionStateNode) Writes() []state.AnyKey      { return []state.AnyKey{n.stateKey} }
func (n *writePositionStateNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *writePositionStateNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = aw
	snapshot := artifact.MustGet(av, n.inputKey)
	state.StageWrite(txn, n.stateKey, snapshot)
	return nil
}

func TestFeatureRangeHighFactoryBuildRequiresWindow(t *testing.T) {
	t.Parallel()

	factory := &FeatureRangeHighFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "range_high",
		Kind:   "feature_range_high",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing window")
	}
}

func TestSignalBreakoutLongFactoryBuildRequiresRangeNodeID(t *testing.T) {
	t.Parallel()

	factory := &SignalBreakoutLongFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "breakout",
		Kind:   "signal_breakout_long",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing range_node_id")
	}
}

func TestFilterSpreadFactoryBuildRequiresMaxSpreadBps(t *testing.T) {
	t.Parallel()

	factory := &FilterSpreadFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:   "spread_filter",
		Kind: "filter_spread",
		Config: map[string]any{
			"allowed_node_id": "session_filter",
		},
	})
	if err == nil {
		t.Fatal("expected validation error for missing max_spread_bps")
	}
}

func TestFeatureATRRangeHighAndBreakoutRun(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}

	atrNode, err := registry.Build(spec.NodeSpec{
		ID:   "atr",
		Kind: "feature_atr",
		Config: map[string]any{
			"window": 3,
		},
	})
	if err != nil {
		t.Fatalf("build atr: %v", err)
	}
	rangeNode, err := registry.Build(spec.NodeSpec{
		ID:   "range",
		Kind: "feature_range_high",
		Config: map[string]any{
			"window": 3,
		},
	})
	if err != nil {
		t.Fatalf("build range high: %v", err)
	}
	breakoutNode, err := registry.Build(spec.NodeSpec{
		ID:   "breakout",
		Kind: "signal_breakout_long",
		Config: map[string]any{
			"range_node_id": "range",
		},
	})
	if err != nil {
		t.Fatalf("build breakout: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, usecase.InputKeyMarketOHLCVBars, []marketdata.OHLCV{
		{
			Open:  marketdata.NewPriceFromRaw(1000),
			High:  marketdata.NewPriceFromRaw(1005),
			Low:   marketdata.NewPriceFromRaw(998),
			Close: marketdata.NewPriceFromRaw(1002),
		},
		{
			Open:  marketdata.NewPriceFromRaw(1002),
			High:  marketdata.NewPriceFromRaw(1006),
			Low:   marketdata.NewPriceFromRaw(1001),
			Close: marketdata.NewPriceFromRaw(1004),
		},
		{
			Open:  marketdata.NewPriceFromRaw(1004),
			High:  marketdata.NewPriceFromRaw(1007),
			Low:   marketdata.NewPriceFromRaw(1003),
			Close: marketdata.NewPriceFromRaw(1005),
		},
		{
			Open:  marketdata.NewPriceFromRaw(1005),
			High:  marketdata.NewPriceFromRaw(1012),
			Low:   marketdata.NewPriceFromRaw(1004),
			Close: marketdata.NewPriceFromRaw(1011),
		},
	})

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	view := artifacts.View()
	if err := atrNode.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run atr: %v", err)
	}
	view = artifacts.View()
	if err := rangeNode.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run range high: %v", err)
	}
	view = artifacts.View()
	if err := breakoutNode.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run breakout: %v", err)
	}

	if got := artifact.MustGet(view, featureATROutputKey("atr")); got.Value.Raw() <= 0 {
		t.Fatalf("expected atr > 0, got %d", got.Value.Raw())
	}
	if got := artifact.MustGet(view, featureRangeHighOutputKey("range")); got.Value.Raw() != 1007 {
		t.Fatalf("expected range high 1007, got %d", got.Value.Raw())
	}
	if got := artifact.MustGet(view, signalBreakoutLongOutputKey("breakout")); !got.Triggered {
		t.Fatal("expected breakout signal true")
	}
}

func TestBreakoutFilterAndRiskRun(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}

	nodes := []spec.NodeSpec{
		{ID: "atr", Kind: "feature_atr", Config: map[string]any{"window": 3}},
		{ID: "range", Kind: "feature_range_high", Config: map[string]any{"window": 3}},
		{ID: "breakout", Kind: "signal_breakout_long", Config: map[string]any{"range_node_id": "range"}},
		{ID: "session_filter", Kind: "filter_session", Config: map[string]any{"signal_node_id": "breakout", "allowed_sessions": []string{"tokyo"}}},
		{ID: "spread_filter", Kind: "filter_spread", Config: map[string]any{"allowed_node_id": "session_filter", "max_spread_bps": 5.0}},
		{ID: "sizing", Kind: "risk_position_sizing", Config: map[string]any{"allowed_node_id": "spread_filter", "atr_node_id": "atr", "risk_rate": 0.01, "stop_atr_multiple": 1.5}},
		{ID: "decision", Kind: "signal_decision_mapper", Config: map[string]any{"allowed_node_id": "spread_filter", "sizing_node_id": "sizing"}},
		{ID: "obs", Kind: "observability_emit_signal_decision", Config: map[string]any{"decision_node_id": "decision"}},
		{ID: "paper_exec", Kind: "execution_submit_paper_order", Config: map[string]any{"decision_node_id": "decision"}},
		{ID: "position", Kind: "position_tracker_update", Config: map[string]any{"execution_node_id": "paper_exec"}},
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, usecase.InputKeyMarketOHLCVBars, []marketdata.OHLCV{
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:00:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
			Open:      marketdata.NewPriceFromRaw(1000),
			High:      marketdata.NewPriceFromRaw(1005),
			Low:       marketdata.NewPriceFromRaw(998),
			Close:     marketdata.NewPriceFromRaw(1002),
		},
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:02:00Z"),
			Open:      marketdata.NewPriceFromRaw(1002),
			High:      marketdata.NewPriceFromRaw(1006),
			Low:       marketdata.NewPriceFromRaw(1001),
			Close:     marketdata.NewPriceFromRaw(1004),
		},
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:02:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:03:00Z"),
			Open:      marketdata.NewPriceFromRaw(1004),
			High:      marketdata.NewPriceFromRaw(1007),
			Low:       marketdata.NewPriceFromRaw(1003),
			Close:     marketdata.NewPriceFromRaw(1005),
		},
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:03:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:04:00Z"),
			Open:      marketdata.NewPriceFromRaw(1005),
			High:      marketdata.NewPriceFromRaw(1012),
			Low:       marketdata.NewPriceFromRaw(1004),
			Close:     marketdata.NewPriceFromRaw(1011),
		},
	})
	artifact.Set(writer, usecase.InputKeyMarketSpreadBps, 1.5)
	artifact.Set(writer, usecase.InputKeyAccountBalance, 10000.0)

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	view := artifacts.View()
	for _, nodeSpec := range nodes {
		built, err := registry.Build(nodeSpec)
		if err != nil {
			t.Fatalf("build node %s: %v", nodeSpec.ID, err)
		}
		if err := built.Run(context.Background(), view, writer, txn); err != nil {
			t.Fatalf("run node %s: %v", nodeSpec.ID, err)
		}
		view = artifacts.View()
	}

	filterResult := artifact.MustGet(view, entryFilterResultKey("spread_filter"))
	if !filterResult.Allowed {
		t.Fatalf("expected allowed=true, got reason=%q", filterResult.Reason)
	}
	if got := artifact.MustGet(view, riskStopLossDistanceOutputKey("sizing")); got.Raw() <= 0 {
		t.Fatalf("expected stop loss distance > 0, got %d", got.Raw())
	}
	if got := artifact.MustGet(view, riskPositionSizeOutputKey("sizing")); got <= 0 {
		t.Fatalf("expected position size > 0, got %f", got)
	}
	decision := artifact.MustGet(view, signalDecisionOutputKey("decision"))
	if decision.Action != "buy" {
		t.Fatalf("expected action=buy, got %q", decision.Action)
	}
	if decision.PositionSize <= 0 {
		t.Fatalf("expected decision position_size > 0, got %f", decision.PositionSize)
	}
	obs := artifact.MustGet(view, observabilitySignalDecisionOutputKey("obs"))
	if !obs.Allowed {
		t.Fatalf("expected observability allowed=true, got %#v", obs)
	}
	execResult := artifact.MustGet(view, paperExecutionResultOutputKey("paper_exec"))
	if !execResult.Submitted {
		t.Fatalf("expected paper execution submitted=true, got %#v", execResult)
	}
	position := artifact.MustGet(view, positionSnapshotOutputKey("position"))
	if !position.HasPosition || position.Side != "long" || position.Size <= 0 {
		t.Fatalf("expected long position snapshot, got %#v", position)
	}
	if position.EntryPrice.Raw() <= 0 {
		t.Fatalf("expected entry price > 0, got %d", position.EntryPrice.Raw())
	}
	riskAmount := 10000.0 * 0.01
	if got := position.Size * float64(artifact.MustGet(view, riskStopLossDistanceOutputKey("sizing")).Raw()); got > riskAmount+1e-9 {
		t.Fatalf("expected risk-constrained sizing, got exposure=%f risk_amount=%f", got, riskAmount)
	}
}

func TestFilterSessionReturnsBlockReason(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	nodes := []spec.NodeSpec{
		{ID: "range", Kind: "feature_range_high", Config: map[string]any{"window": 2}},
		{ID: "breakout", Kind: "signal_breakout_long", Config: map[string]any{"range_node_id": "range"}},
		{ID: "session_filter", Kind: "filter_session", Config: map[string]any{"signal_node_id": "breakout", "allowed_sessions": []string{"newyork"}}},
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, usecase.InputKeyMarketOHLCVBars, []marketdata.OHLCV{
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:00:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
			Open:      marketdata.NewPriceFromRaw(1000),
			High:      marketdata.NewPriceFromRaw(1005),
			Low:       marketdata.NewPriceFromRaw(998),
			Close:     marketdata.NewPriceFromRaw(1002),
		},
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:02:00Z"),
			Open:      marketdata.NewPriceFromRaw(1002),
			High:      marketdata.NewPriceFromRaw(1006),
			Low:       marketdata.NewPriceFromRaw(1001),
			Close:     marketdata.NewPriceFromRaw(1004),
		},
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:02:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:03:00Z"),
			Open:      marketdata.NewPriceFromRaw(1004),
			High:      marketdata.NewPriceFromRaw(1012),
			Low:       marketdata.NewPriceFromRaw(1003),
			Close:     marketdata.NewPriceFromRaw(1011),
		},
	})

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	view := artifacts.View()
	for _, nodeSpec := range nodes {
		built, err := registry.Build(nodeSpec)
		if err != nil {
			t.Fatalf("build node %s: %v", nodeSpec.ID, err)
		}
		if err := built.Run(context.Background(), view, writer, txn); err != nil {
			t.Fatalf("run node %s: %v", nodeSpec.ID, err)
		}
		view = artifacts.View()
	}

	got := artifact.MustGet(view, entryFilterResultKey("session_filter"))
	if got.Allowed {
		t.Fatalf("expected allowed=false, got %#v", got)
	}
	if got.Reason != "session_blocked" {
		t.Fatalf("expected reason=session_blocked, got %q", got.Reason)
	}
}

func TestFilterSpreadReturnsBlockReason(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	nodes := []spec.NodeSpec{
		{ID: "range", Kind: "feature_range_high", Config: map[string]any{"window": 2}},
		{ID: "breakout", Kind: "signal_breakout_long", Config: map[string]any{"range_node_id": "range"}},
		{ID: "session_filter", Kind: "filter_session", Config: map[string]any{"signal_node_id": "breakout", "allowed_sessions": []string{"tokyo"}}},
		{ID: "spread_filter", Kind: "filter_spread", Config: map[string]any{"allowed_node_id": "session_filter", "max_spread_bps": 1.0}},
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, usecase.InputKeyMarketOHLCVBars, []marketdata.OHLCV{
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:00:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
			Open:      marketdata.NewPriceFromRaw(1000),
			High:      marketdata.NewPriceFromRaw(1005),
			Low:       marketdata.NewPriceFromRaw(998),
			Close:     marketdata.NewPriceFromRaw(1002),
		},
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:02:00Z"),
			Open:      marketdata.NewPriceFromRaw(1002),
			High:      marketdata.NewPriceFromRaw(1006),
			Low:       marketdata.NewPriceFromRaw(1001),
			Close:     marketdata.NewPriceFromRaw(1004),
		},
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:02:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:03:00Z"),
			Open:      marketdata.NewPriceFromRaw(1004),
			High:      marketdata.NewPriceFromRaw(1012),
			Low:       marketdata.NewPriceFromRaw(1003),
			Close:     marketdata.NewPriceFromRaw(1011),
		},
	})
	artifact.Set(writer, usecase.InputKeyMarketSpreadBps, 2.0)

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	view := artifacts.View()
	for _, nodeSpec := range nodes {
		built, err := registry.Build(nodeSpec)
		if err != nil {
			t.Fatalf("build node %s: %v", nodeSpec.ID, err)
		}
		if err := built.Run(context.Background(), view, writer, txn); err != nil {
			t.Fatalf("run node %s: %v", nodeSpec.ID, err)
		}
		view = artifacts.View()
	}

	got := artifact.MustGet(view, entryFilterResultKey("spread_filter"))
	if got.Allowed {
		t.Fatalf("expected allowed=false, got %#v", got)
	}
	if got.Reason != "spread_limit" {
		t.Fatalf("expected reason=spread_limit, got %q", got.Reason)
	}
}

func TestSignalDecisionAndObservabilityWithoutExecutionNodes(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	nodes := []spec.NodeSpec{
		{ID: "atr", Kind: "feature_atr", Config: map[string]any{"window": 3}},
		{ID: "range", Kind: "feature_range_high", Config: map[string]any{"window": 3}},
		{ID: "breakout", Kind: "signal_breakout_long", Config: map[string]any{"range_node_id": "range"}},
		{ID: "session_filter", Kind: "filter_session", Config: map[string]any{"signal_node_id": "breakout", "allowed_sessions": []string{"tokyo"}}},
		{ID: "spread_filter", Kind: "filter_spread", Config: map[string]any{"allowed_node_id": "session_filter", "max_spread_bps": 5.0}},
		{ID: "sizing", Kind: "risk_position_sizing", Config: map[string]any{"allowed_node_id": "spread_filter", "atr_node_id": "atr", "risk_rate": 0.01, "stop_atr_multiple": 1.5}},
		{ID: "decision", Kind: "signal_decision_mapper", Config: map[string]any{"allowed_node_id": "spread_filter", "sizing_node_id": "sizing"}},
		{ID: "obs", Kind: "observability_emit_signal_decision", Config: map[string]any{"decision_node_id": "decision"}},
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, usecase.InputKeyMarketOHLCVBars, []marketdata.OHLCV{
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:00:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
			Open:      marketdata.NewPriceFromRaw(1000),
			High:      marketdata.NewPriceFromRaw(1005),
			Low:       marketdata.NewPriceFromRaw(998),
			Close:     marketdata.NewPriceFromRaw(1002),
		},
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:02:00Z"),
			Open:      marketdata.NewPriceFromRaw(1002),
			High:      marketdata.NewPriceFromRaw(1006),
			Low:       marketdata.NewPriceFromRaw(1001),
			Close:     marketdata.NewPriceFromRaw(1004),
		},
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:02:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:03:00Z"),
			Open:      marketdata.NewPriceFromRaw(1004),
			High:      marketdata.NewPriceFromRaw(1007),
			Low:       marketdata.NewPriceFromRaw(1003),
			Close:     marketdata.NewPriceFromRaw(1005),
		},
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:03:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:04:00Z"),
			Open:      marketdata.NewPriceFromRaw(1005),
			High:      marketdata.NewPriceFromRaw(1012),
			Low:       marketdata.NewPriceFromRaw(1004),
			Close:     marketdata.NewPriceFromRaw(1011),
		},
	})
	artifact.Set(writer, usecase.InputKeyMarketSpreadBps, 1.5)
	artifact.Set(writer, usecase.InputKeyAccountBalance, 10000.0)

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	view := artifacts.View()
	for _, nodeSpec := range nodes {
		built, err := registry.Build(nodeSpec)
		if err != nil {
			t.Fatalf("build node %s: %v", nodeSpec.ID, err)
		}
		if err := built.Run(context.Background(), view, writer, txn); err != nil {
			t.Fatalf("run node %s: %v", nodeSpec.ID, err)
		}
		view = artifacts.View()
	}

	decision := artifact.MustGet(view, signalDecisionOutputKey("decision"))
	if decision.Action != "buy" {
		t.Fatalf("expected action=buy, got %q", decision.Action)
	}
	obs := artifact.MustGet(view, observabilitySignalDecisionOutputKey("obs"))
	if !obs.Allowed {
		t.Fatalf("expected allowed=true, got %#v", obs)
	}
}

func TestBreakoutArtifactStableIDsRegression(t *testing.T) {
	t.Parallel()

	if got := featureATROutputKey("atr").Raw().StableID; got != "artifact:dagruntime.feature_atr.atr.v1" {
		t.Fatalf("unexpected feature_atr stable id: %s", got)
	}
	if got := signalDecisionOutputKey("decision").Raw().StableID; got != "artifact:dagruntime.signal.decision.decision.v1" {
		t.Fatalf("unexpected signal decision stable id: %s", got)
	}
	if got := positionSnapshotOutputKey("position").Raw().StableID; got != "artifact:dagruntime.position.snapshot.position.v1" {
		t.Fatalf("unexpected position snapshot stable id: %s", got)
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
