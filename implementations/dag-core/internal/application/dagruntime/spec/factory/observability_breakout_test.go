package factory

import (
	"context"
	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/algotrade"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/marketdata"
	artifactinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
	"testing"
)

func TestObservabilityEmitOrderDecisionFactoryBuildRequiresDecisionNodeID(t *testing.T) {
	t.Parallel()

	factory := &ObservabilityEmitOrderDecisionFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "obs_order",
		Kind:   "observability_emit_order_decision",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing decision_node_id")
	}
}

func TestObservabilityEmitOrderDecisionRun(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	obsNode, err := registry.Build(spec.NodeSpec{
		ID:   "obs_order",
		Kind: "observability_emit_order_decision",
		Config: map[string]any{
			"decision_node_id": "decision",
		},
	})
	if err != nil {
		t.Fatalf("build observability_emit_order_decision: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, signalDecisionOutputKey("decision"), algotrade.OrderRequest{
		Action:           algotrade.OrderActionBuy,
		Reason:           "entry_allowed",
		PositionSize:     0.9,
		StopLossDistance: marketdata.NewPriceFromRaw(12),
	})
	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	if err := obsNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run observability_emit_order_decision: %v", err)
	}
	got := artifact.MustGet(artifacts.View(), observabilityOrderDecisionOutputKey("obs_order"))
	if got.Action != algotrade.OrderActionBuy || got.PositionSize <= 0 {
		t.Fatalf("unexpected order decision observability payload: %#v", got)
	}
}

func TestObservabilityEmitPositionEventFactoryBuildRequiresPositionNodeID(t *testing.T) {
	t.Parallel()

	factory := &ObservabilityEmitPositionEventFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "obs_position",
		Kind:   "observability_emit_position_event",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing position_node_id")
	}
}

func TestObservabilityEmitPositionEventRun(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	obsNode, err := registry.Build(spec.NodeSpec{
		ID:   "obs_position",
		Kind: "observability_emit_position_event",
		Config: map[string]any{
			"position_node_id": "position_load",
		},
	})
	if err != nil {
		t.Fatalf("build observability_emit_position_event: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, positionSnapshotOutputKey("position_load"), algotrade.PositionSnapshot{
		HasPosition: true,
		Side:        algotrade.PositionSideLong,
		Size:        0.4,
		EntryPrice:  marketdata.NewPriceFromRaw(1000),
	})
	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	if err := obsNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run observability_emit_position_event: %v", err)
	}
	got := artifact.MustGet(artifacts.View(), observabilityPositionEventOutputKey("obs_position"))
	if !got.HasPosition || got.Side != algotrade.PositionSideLong || got.Size <= 0 {
		t.Fatalf("unexpected position event observability payload: %#v", got)
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
