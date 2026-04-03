package factory

import (
	"context"
	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/algotrade"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/marketdata"
	artifactinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
	"testing"
)

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
	rangeLowNode, err := registry.Build(spec.NodeSpec{
		ID:   "range_low",
		Kind: "feature_range_low",
		Config: map[string]any{
			"window": 3,
		},
	})
	if err != nil {
		t.Fatalf("build range low: %v", err)
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
	if err := rangeLowNode.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run range low: %v", err)
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
	if got := artifact.MustGet(view, featureRangeLowOutputKey("range_low")); got.Value.Raw() != 998 {
		t.Fatalf("expected range low 998, got %d", got.Value.Raw())
	}
	if got := artifact.MustGet(view, signalBreakoutLongOutputKey("breakout")); !got.Triggered {
		t.Fatal("expected breakout signal true")
	}
}

func TestFeatureHigherTFTrendRun(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	nodeBuilt, err := registry.Build(spec.NodeSpec{
		ID:   "htf",
		Kind: "feature_higher_tf_trend",
	})
	if err != nil {
		t.Fatalf("build htf trend: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, inputKeyMarketOHLCVBarsH1, []marketdata.OHLCV{
		{
			Open:  marketdata.NewPriceFromRaw(1000),
			Close: marketdata.NewPriceFromRaw(1002),
		},
		{
			Open:  marketdata.NewPriceFromRaw(1002),
			Close: marketdata.NewPriceFromRaw(1010),
		},
	})
	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	if err := nodeBuilt.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run htf trend: %v", err)
	}
	got := artifact.MustGet(artifacts.View(), featureHigherTFTrendOutputKey("htf"))
	if got.Direction != "up" {
		t.Fatalf("expected direction=up, got %s", got.Direction)
	}
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

func TestFeatureRangeLowAndBreakoutShortRun(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}

	rangeLowNode, err := registry.Build(spec.NodeSpec{
		ID:   "range_low",
		Kind: "feature_range_low",
		Config: map[string]any{
			"window": 3,
		},
	})
	if err != nil {
		t.Fatalf("build range low: %v", err)
	}
	breakoutShortNode, err := registry.Build(spec.NodeSpec{
		ID:   "breakout_short",
		Kind: "signal_breakout_short",
		Config: map[string]any{
			"range_node_id": "range_low",
		},
	})
	if err != nil {
		t.Fatalf("build breakout short: %v", err)
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
			High:  marketdata.NewPriceFromRaw(1006),
			Low:   marketdata.NewPriceFromRaw(990),
			Close: marketdata.NewPriceFromRaw(992),
		},
	})

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	view := artifacts.View()
	if err := rangeLowNode.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run range low: %v", err)
	}
	view = artifacts.View()
	if err := breakoutShortNode.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run breakout short: %v", err)
	}
	view = artifacts.View()

	if got := artifact.MustGet(view, featureRangeLowOutputKey("range_low")); got.Value.Raw() != 998 {
		t.Fatalf("expected range low 998, got %d", got.Value.Raw())
	}
	if got := artifact.MustGet(view, signalBreakoutShortOutputKey("breakout_short")); !got.Triggered {
		t.Fatal("expected breakout short signal true")
	}
}

func TestFeatureRangeLowFactoryBuildRequiresWindow(t *testing.T) {
	t.Parallel()

	factory := &FeatureRangeLowFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "range_low",
		Kind:   "feature_range_low",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing window")
	}
}

func TestFeatureSpreadAndSessionStateRun(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	spreadNode, err := registry.Build(spec.NodeSpec{
		ID:   "spread",
		Kind: "feature_spread",
	})
	if err != nil {
		t.Fatalf("build spread: %v", err)
	}
	sessionNode, err := registry.Build(spec.NodeSpec{
		ID:   "session",
		Kind: "feature_session_state",
	})
	if err != nil {
		t.Fatalf("build session state: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, usecase.InputKeyMarketSpreadBps, 1.8)
	artifact.Set(writer, usecase.InputKeyMarketOHLCVBars, []marketdata.OHLCV{
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T07:00:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T07:01:00Z"),
			Open:      marketdata.NewPriceFromRaw(1000),
			High:      marketdata.NewPriceFromRaw(1005),
			Low:       marketdata.NewPriceFromRaw(999),
			Close:     marketdata.NewPriceFromRaw(1002),
		},
	})

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	view := artifacts.View()
	if err := spreadNode.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run spread: %v", err)
	}
	view = artifacts.View()
	if err := sessionNode.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run session state: %v", err)
	}
	view = artifacts.View()

	if got := artifact.MustGet(view, featureSpreadOutputKey("spread")); got.Bps != 1.8 {
		t.Fatalf("expected spread=1.8, got %f", got.Bps)
	}
	if got := artifact.MustGet(view, featureSessionStateOutputKey("session")); got.Session != "tokyo" {
		t.Fatalf("expected session=tokyo, got %s", got.Session)
	}
}

func TestFilterDailyLossLimitFactoryBuildRequiresAllowedNodeID(t *testing.T) {
	t.Parallel()

	factory := &FilterDailyLossLimitFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "daily_loss_filter",
		Kind:   "filter_daily_loss_limit",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing allowed_node_id")
	}
}

func TestFilterDailyLossLimitRun(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	filterNode, err := registry.Build(spec.NodeSpec{
		ID:   "daily_loss_filter",
		Kind: "filter_daily_loss_limit",
		Config: map[string]any{
			"allowed_node_id": "session_filter",
			"max_daily_loss":  100.0,
		},
	})
	if err != nil {
		t.Fatalf("build filter_daily_loss_limit: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, entryFilterResultKey("session_filter"), algotrade.EntryFilterResult{
		Allowed: true,
	})

	mem := stateinfra.NewMemoryStore()
	txn := mem.BeginTxn("test")
	state.StageWrite(txn, algotrade.StateDailyPnL, algotrade.DailyPnLState{
		TradingDay:   "2026-04-01",
		RealizedPnL:  -120,
		LossLimitHit: false,
	})

	view := artifacts.View()
	if err := filterNode.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run filter_daily_loss_limit: %v", err)
	}
	got := artifact.MustGet(artifacts.View(), entryFilterResultKey("daily_loss_filter"))
	if got.Allowed {
		t.Fatalf("expected allowed=false, got %#v", got)
	}
	if got.Reason != "daily_loss_limit" {
		t.Fatalf("expected reason=daily_loss_limit, got %q", got.Reason)
	}
}

func TestFilterEconomicEventFactoryBuildRequiresAllowedNodeID(t *testing.T) {
	t.Parallel()

	factory := &FilterEconomicEventFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "econ_filter",
		Kind:   "filter_economic_event",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing allowed_node_id")
	}
}

func TestFilterEconomicEventRun(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	filterNode, err := registry.Build(spec.NodeSpec{
		ID:   "econ_filter",
		Kind: "filter_economic_event",
		Config: map[string]any{
			"allowed_node_id": "session_filter",
			"block_severity":  "high",
		},
	})
	if err != nil {
		t.Fatalf("build filter_economic_event: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, entryFilterResultKey("session_filter"), algotrade.EntryFilterResult{Allowed: true})
	artifact.Set(writer, inputKeyEconomicEvents, []algotrade.EconomicEvent{
		{Name: "CPI", Severity: "high"},
	})

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	if err := filterNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run filter_economic_event: %v", err)
	}
	got := artifact.MustGet(artifacts.View(), entryFilterResultKey("econ_filter"))
	if got.Allowed {
		t.Fatalf("expected allowed=false, got %#v", got)
	}
	if got.Reason != "economic_event" {
		t.Fatalf("expected reason=economic_event, got %q", got.Reason)
	}
}

func TestFilterHigherTFAlignmentFactoryBuildRequiresDependencies(t *testing.T) {
	t.Parallel()

	factory := &FilterHigherTFAlignmentFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:   "htf_filter",
		Kind: "filter_higher_tf_alignment",
		Config: map[string]any{
			"allowed_node_id": "session_filter",
		},
	})
	if err == nil {
		t.Fatal("expected validation error for missing trend_node_id")
	}
}

func TestFilterHigherTFAlignmentRun(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	filterNode, err := registry.Build(spec.NodeSpec{
		ID:   "htf_filter",
		Kind: "filter_higher_tf_alignment",
		Config: map[string]any{
			"allowed_node_id":    "session_filter",
			"trend_node_id":      "htf",
			"required_direction": "up",
		},
	})
	if err != nil {
		t.Fatalf("build filter_higher_tf_alignment: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, entryFilterResultKey("session_filter"), algotrade.EntryFilterResult{Allowed: true})
	artifact.Set(writer, featureHigherTFTrendOutputKey("htf"), algotrade.FeatureHigherTFTrend{
		Direction: "down",
	})

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	if err := filterNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run filter_higher_tf_alignment: %v", err)
	}
	got := artifact.MustGet(artifacts.View(), entryFilterResultKey("htf_filter"))
	if got.Allowed {
		t.Fatalf("expected allowed=false, got %#v", got)
	}
	if got.Reason != "higher_tf_alignment" {
		t.Fatalf("expected reason=higher_tf_alignment, got %q", got.Reason)
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

func TestMarketBarInputH1FactoryBuildRejectsUnknownConfig(t *testing.T) {
	t.Parallel()

	factory := &MarketBarInputH1Factory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:   "bar_h1_input",
		Kind: "market_bar_input_h1",
		Config: map[string]any{
			"unexpected": true,
		},
	})
	if err == nil {
		t.Fatal("expected unknown config key error")
	}
}

func TestMarketBarInputM1FactoryBuildRejectsUnknownConfig(t *testing.T) {
	t.Parallel()

	factory := &MarketBarInputM1Factory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:   "bar_m1_input",
		Kind: "market_bar_input_m1",
		Config: map[string]any{
			"unexpected": true,
		},
	})
	if err == nil {
		t.Fatal("expected unknown config key error")
	}
}

func TestMarketInputBridgeNodesRun(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	tickNode, err := registry.Build(spec.NodeSpec{
		ID:   "tick_input",
		Kind: "market_tick_input",
	})
	if err != nil {
		t.Fatalf("build market_tick_input: %v", err)
	}
	barM1Node, err := registry.Build(spec.NodeSpec{
		ID:   "bar_m1_input",
		Kind: "market_bar_input_m1",
	})
	if err != nil {
		t.Fatalf("build market_bar_input_m1: %v", err)
	}
	barH1Node, err := registry.Build(spec.NodeSpec{
		ID:   "bar_h1_input",
		Kind: "market_bar_input_h1",
	})
	if err != nil {
		t.Fatalf("build market_bar_input_h1: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, usecase.InputKeyMarketTick, marketdata.Tick{
		SymbolID: 1,
		Time:     marketdata.MustParseUTCTime("2026-04-02T00:00:00Z"),
		Bid:      marketdata.NewPriceFromRaw(1000),
		Ask:      marketdata.NewPriceFromRaw(1002),
	})
	artifact.Set(writer, usecase.InputKeyMarketOHLCVBars, []marketdata.OHLCV{
		{
			Open:  marketdata.NewPriceFromRaw(1000),
			High:  marketdata.NewPriceFromRaw(1005),
			Low:   marketdata.NewPriceFromRaw(998),
			Close: marketdata.NewPriceFromRaw(1004),
		},
	})
	artifact.Set(writer, inputKeyMarketOHLCVBarsH1, []marketdata.OHLCV{
		{
			Open:  marketdata.NewPriceFromRaw(990),
			High:  marketdata.NewPriceFromRaw(1010),
			Low:   marketdata.NewPriceFromRaw(980),
			Close: marketdata.NewPriceFromRaw(1008),
		},
	})

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	view := artifacts.View()
	if err := tickNode.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run market_tick_input: %v", err)
	}
	view = artifacts.View()
	if err := barM1Node.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run market_bar_input_m1: %v", err)
	}
	view = artifacts.View()
	if err := barH1Node.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run market_bar_input_h1: %v", err)
	}
	view = artifacts.View()

	if got := artifact.MustGet(view, marketTickInputOutputKey("tick_input")); got.Ask.Raw() != 1002 {
		t.Fatalf("expected bridged tick ask=1002, got %d", got.Ask.Raw())
	}
	if got := artifact.MustGet(view, marketBarInputM1OutputKey("bar_m1_input")); len(got) != 1 {
		t.Fatalf("expected bridged m1 bars len=1, got %d", len(got))
	}
	if got := artifact.MustGet(view, marketBarInputH1OutputKey("bar_h1_input")); len(got) != 1 {
		t.Fatalf("expected bridged h1 bars len=1, got %d", len(got))
	}
}

func TestMarketTickInputFactoryBuildRejectsUnknownConfig(t *testing.T) {
	t.Parallel()

	factory := &MarketTickInputFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:   "tick_input",
		Kind: "market_tick_input",
		Config: map[string]any{
			"unexpected": true,
		},
	})
	if err == nil {
		t.Fatal("expected unknown config key error")
	}
}

func TestRiskMaxPositionsCheckFactoryBuildRequiresAllowedNodeID(t *testing.T) {
	t.Parallel()

	factory := &RiskMaxPositionsCheckFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "max_positions",
		Kind:   "risk_max_positions_check",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing allowed_node_id")
	}
}

func TestRiskMaxPositionsCheckRun(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	checkNode, err := registry.Build(spec.NodeSpec{
		ID:   "max_positions",
		Kind: "risk_max_positions_check",
		Config: map[string]any{
			"allowed_node_id": "session_filter",
			"max_positions":   1,
		},
	})
	if err != nil {
		t.Fatalf("build risk_max_positions_check: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, entryFilterResultKey("session_filter"), algotrade.EntryFilterResult{
		Allowed: true,
	})

	mem := stateinfra.NewMemoryStore()
	txn := mem.BeginTxn("test")
	state.StageWrite(txn, algotrade.StateOpenPositions, algotrade.OpenPositionsState{
		Items: []algotrade.OpenPosition{
			{
				PositionID: "p1",
				Symbol:     "USDJPY",
			},
		},
	})

	view := artifacts.View()
	if err := checkNode.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run risk_max_positions_check: %v", err)
	}
	got := artifact.MustGet(artifacts.View(), entryFilterResultKey("max_positions"))
	if got.Allowed {
		t.Fatalf("expected allowed=false, got %#v", got)
	}
	if got.Reason != "max_positions" {
		t.Fatalf("expected reason=max_positions, got %q", got.Reason)
	}
}

func TestRiskStopLossAndTakeProfitRun(t *testing.T) {
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
	slNode, err := registry.Build(spec.NodeSpec{
		ID:   "sl_from_atr",
		Kind: "risk_stop_loss_from_atr",
		Config: map[string]any{
			"atr_node_id":  "atr",
			"atr_multiple": 1.5,
		},
	})
	if err != nil {
		t.Fatalf("build sl_from_atr: %v", err)
	}
	tpNode, err := registry.Build(spec.NodeSpec{
		ID:   "tp_from_rr",
		Kind: "risk_take_profit_from_rr",
		Config: map[string]any{
			"stop_loss_node_id": "sl_from_atr",
			"rr_ratio":          2.0,
		},
	})
	if err != nil {
		t.Fatalf("build tp_from_rr: %v", err)
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
	if err := slNode.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run sl_from_atr: %v", err)
	}
	view = artifacts.View()
	if err := tpNode.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run tp_from_rr: %v", err)
	}
	view = artifacts.View()

	sl := artifact.MustGet(view, riskStopLossDistanceOutputKey("sl_from_atr"))
	if sl.Raw() <= 0 {
		t.Fatalf("expected stop loss distance > 0, got %d", sl.Raw())
	}
	tp := artifact.MustGet(view, riskTakeProfitDistanceOutputKey("tp_from_rr"))
	if tp.Raw() != sl.Raw()*2 {
		t.Fatalf("expected take profit distance = 2x stop loss, got sl=%d tp=%d", sl.Raw(), tp.Raw())
	}
}

func TestRiskStopLossFromATRFactoryBuildRequiresATRNodeID(t *testing.T) {
	t.Parallel()

	factory := &RiskStopLossFromATRFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "sl_from_atr",
		Kind:   "risk_stop_loss_from_atr",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing atr_node_id")
	}
}

func TestRiskTakeProfitFromRRFactoryBuildRequiresStopLossNodeID(t *testing.T) {
	t.Parallel()

	factory := &RiskTakeProfitFromRRFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:   "tp_from_rr",
		Kind: "risk_take_profit_from_rr",
		Config: map[string]any{
			"rr_ratio": 2.0,
		},
	})
	if err == nil {
		t.Fatal("expected validation error for missing stop_loss_node_id")
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

func TestSignalBreakoutShortFactoryBuildRequiresRangeNodeID(t *testing.T) {
	t.Parallel()

	factory := &SignalBreakoutShortFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "breakout_short",
		Kind:   "signal_breakout_short",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing range_node_id")
	}
}

func TestSignalExitBasicFactoryBuildRequiresDependencies(t *testing.T) {
	t.Parallel()

	factory := &SignalExitBasicFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "exit",
		Kind:   "signal_exit_basic",
		Config: map[string]any{"position_node_id": "position"},
	})
	if err == nil {
		t.Fatal("expected validation error for missing stop_loss_node_id")
	}
}

func TestSignalExitBasicRun(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	exitNode, err := registry.Build(spec.NodeSpec{
		ID:   "exit",
		Kind: "signal_exit_basic",
		Config: map[string]any{
			"position_node_id":  "position",
			"stop_loss_node_id": "sizing",
		},
	})
	if err != nil {
		t.Fatalf("build exit node: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, positionSnapshotOutputKey("position"), algotrade.PositionSnapshot{
		HasPosition: true,
		Side:        algotrade.PositionSideLong,
		Size:        1.0,
		EntryPrice:  marketdata.NewPriceFromRaw(1000),
	})
	artifact.Set(writer, riskStopLossDistanceOutputKey("sizing"), marketdata.NewPriceFromRaw(50))
	artifact.Set(writer, usecase.InputKeyMarketOHLCVBars, []marketdata.OHLCV{
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:00:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
			Open:      marketdata.NewPriceFromRaw(980),
			High:      marketdata.NewPriceFromRaw(982),
			Low:       marketdata.NewPriceFromRaw(948),
			Close:     marketdata.NewPriceFromRaw(949),
		},
	})

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	view := artifacts.View()
	if err := exitNode.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run exit node: %v", err)
	}
	view = artifacts.View()
	exitDecision := artifact.MustGet(view, signalExitDecisionOutputKey("exit"))
	if !exitDecision.ShouldExit {
		t.Fatalf("expected should_exit=true, got %#v", exitDecision)
	}
	if exitDecision.Reason != "stop_loss_hit" {
		t.Fatalf("expected reason=stop_loss_hit, got %q", exitDecision.Reason)
	}
}
