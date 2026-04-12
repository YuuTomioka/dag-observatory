package factory

import (
	"context"
	"strings"
	"testing"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	artifactinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
)

func TestNewBuiltinRegistryRegistersKinds(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	for _, kind := range []string{
		"heavy_calc",
		"sma",
		"cross_detector",
		"signal_mapper",
		"market_tick_input",
		"market_bar_input_m1",
		"market_bar_input_h1",
		"feature_atr",
		"feature_range_high",
		"feature_range_low",
		"feature_spread",
		"feature_session_state",
		"feature_higher_tf_trend",
		"signal_breakout_long",
		"signal_breakout_short",
		"signal_exit_basic",
		"filter_session",
		"filter_spread",
		"filter_economic_event",
		"filter_higher_tf_alignment",
		"filter_daily_loss_limit",
		"risk_position_sizing",
		"risk_max_positions_check",
		"risk_stop_loss_from_atr",
		"risk_take_profit_from_rr",
		"signal_decision_mapper",
		"observability_emit_signal_decision",
		"observability_emit_order_decision",
		"observability_emit_position_event",
		"execution_submit_paper_order",
		"execution_submit_market_order",
		"execution_confirm_fill",
		"position_snapshot_load",
		"position_breakeven",
		"position_trailing_stop",
		"position_timeout_exit",
		"position_tracker_update",
		"position_close_to_closed_trade",
		"closed_trade_store",
		"daily_pnl_update",
		"open_position_close",
		"strategy_summary_update",
	} {
		if !registry.Has(kind) {
			t.Fatalf("expected kind %q to be registered", kind)
		}
	}
}

func TestSMAFactoryBuildRequiresWindow(t *testing.T) {
	t.Parallel()

	factory := &SMAFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "fast_sma",
		Kind:   "sma",
		Config: map[string]any{"source": "close"},
	})
	if err == nil {
		t.Fatal("expected validation error for missing window")
	}
}

func TestSignalMapperFactoryUnknownConfigKey(t *testing.T) {
	t.Parallel()

	factory := &SignalMapperFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID: "emit",
		Config: map[string]any{
			"cross_node_id": "cross",
			"unknown":       "x",
		},
	})
	if err == nil {
		t.Fatal("expected unknown config key error")
	}
}

func TestSMAAndCrossAndSignalRun(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}

	smaFast, err := registry.Build(spec.NodeSpec{
		ID:   "fast_sma",
		Kind: "sma",
		Config: map[string]any{
			"window": 2,
			"source": "close",
		},
	})
	if err != nil {
		t.Fatalf("build fast sma: %v", err)
	}
	smaSlow, err := registry.Build(spec.NodeSpec{
		ID:   "slow_sma",
		Kind: "sma",
		Config: map[string]any{
			"window": 3,
			"source": "close",
		},
	})
	if err != nil {
		t.Fatalf("build slow sma: %v", err)
	}
	cross, err := registry.Build(spec.NodeSpec{
		ID:   "cross",
		Kind: "cross_detector",
		Config: map[string]any{
			"fast_node_id": "fast_sma",
			"slow_node_id": "slow_sma",
			"mode":         "both",
		},
	})
	if err != nil {
		t.Fatalf("build cross: %v", err)
	}
	mapper, err := registry.Build(spec.NodeSpec{
		ID:   "emit",
		Kind: "signal_mapper",
		Config: map[string]any{
			"cross_node_id": "cross",
			"on_cross_up":   "buy",
			"on_cross_down": "sell",
			"on_no_cross":   "hold",
		},
	})
	if err != nil {
		t.Fatalf("build mapper: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, usecase.InputKeyMarketBars, []float64{1.0, 2.0, 4.0})
	artifact.Set(writer, usecase.InputKeySymbol, "USDJPY")
	artifact.Set(writer, usecase.InputKeyMode, "normal")

	txn := stateinfra.NewMemoryStore().BeginTxn("test")

	view := artifacts.View()
	if err := smaFast.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run fast sma: %v", err)
	}
	view = artifacts.View()
	if err := smaSlow.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run slow sma: %v", err)
	}
	view = artifacts.View()
	if err := cross.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run cross: %v", err)
	}
	view = artifacts.View()
	if err := mapper.Run(context.Background(), view, writer, txn); err != nil {
		t.Fatalf("run mapper: %v", err)
	}
	if got := artifact.MustGet(view, signalOutputKey("emit")); got != "buy" {
		t.Fatalf("expected buy signal, got %q", got)
	}
}

func TestHeavyCalcRunMissingRequiredInputReturnsError(t *testing.T) {
	t.Parallel()

	node := &heavyCalcNode{}
	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	txn := stateinfra.NewMemoryStore().BeginTxn("test")

	err := node.Run(context.Background(), artifacts.View(), writer, txn)
	if err == nil {
		t.Fatal("expected error for missing required input")
	}
	if !strings.Contains(err.Error(), usecase.InputKeySymbol.String()) {
		t.Fatalf("expected missing symbol key error, got %v", err)
	}
}
