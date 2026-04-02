package factory

import (
	"context"
	"testing"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/marketdata"
	artifactinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
)

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

	if got := artifact.MustGet(view, featureATROutputKey("atr")); got.Raw() <= 0 {
		t.Fatalf("expected atr > 0, got %d", got.Raw())
	}
	if got := artifact.MustGet(view, featureRangeHighOutputKey("range")); got.Raw() != 1007 {
		t.Fatalf("expected range high 1007, got %d", got.Raw())
	}
	if got := artifact.MustGet(view, signalBreakoutLongOutputKey("breakout")); !got {
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
}
