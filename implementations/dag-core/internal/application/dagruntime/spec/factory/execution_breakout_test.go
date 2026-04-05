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

func TestDecisionOrderPositionArtifactTraceRegression(t *testing.T) {
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
		{ID: "market_exec", Kind: "execution_submit_market_order", Config: map[string]any{"decision_node_id": "decision"}},
		{ID: "position_load", Kind: "position_snapshot_load"},
		{ID: "obs_order", Kind: "observability_emit_order_decision", Config: map[string]any{"decision_node_id": "decision"}},
		{ID: "obs_position", Kind: "observability_emit_position_event", Config: map[string]any{"position_node_id": "position_load"}},
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, usecase.InputKeySymbol, "USDJPY")
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
	artifact.Set(writer, usecase.InputKeyMarketSpreadBps, 1.2)
	artifact.Set(writer, usecase.InputKeyAccountBalance, 10000.0)

	mem := stateinfra.NewMemoryStore()
	txn := mem.BeginTxn("test")
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

	if !artifacts.Has(tradeIntentOutputKey("decision")) {
		t.Fatal("expected trade intent artifact")
	}
	if !artifacts.Has(executionResultOutputKey("market_exec")) {
		t.Fatal("expected execution result artifact")
	}
	if !artifacts.Has(executionResultOutputKey("fill_confirm")) && artifacts.Has(executionFillResultOutputKey("fill_confirm")) {
		t.Fatal("expected fill-side execution result artifact when fill artifact exists")
	}
	if !artifacts.Has(executionMarketOrderRequestOutputKey("market_exec")) {
		t.Fatal("expected market order request artifact")
	}
	if !artifacts.Has(positionSnapshotOutputKey("position_load")) {
		t.Fatal("expected position snapshot artifact")
	}
	if !artifacts.Has(observabilityOrderDecisionOutputKey("obs_order")) {
		t.Fatal("expected order decision observability artifact")
	}
	if !artifacts.Has(observabilityPositionEventOutputKey("obs_position")) {
		t.Fatal("expected position event observability artifact")
	}
}

func TestExecutionConfirmFillFactoryBuildRequiresOrderRequestNodeID(t *testing.T) {
	t.Parallel()

	factory := &ExecutionConfirmFillFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "fill_confirm",
		Kind:   "execution_confirm_fill",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing order_request_node_id")
	}
}

func TestExecutionConfirmFillRunMovesPendingToOpenPosition(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	confirmNode, err := registry.Build(spec.NodeSpec{
		ID:   "fill_confirm",
		Kind: "execution_confirm_fill",
		Config: map[string]any{
			"order_request_node_id": "market_exec",
		},
	})
	if err != nil {
		t.Fatalf("build execution_confirm_fill: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, executionMarketOrderRequestOutputKey("market_exec"), algotrade.MarketOrderRequest{
		Submitted: true,
		OrderID:   "market_exec-1",
		Action:    algotrade.OrderActionBuy,
		Size:      0.7,
		Reason:    "submitted",
		SourceID:  "buy:entry_allowed:0.70000000:20",
	})
	artifact.Set(writer, usecase.InputKeyMarketOHLCVBars, []marketdata.OHLCV{
		{
			Open:  marketdata.NewPriceFromRaw(1000),
			High:  marketdata.NewPriceFromRaw(1008),
			Low:   marketdata.NewPriceFromRaw(998),
			Close: marketdata.NewPriceFromRaw(1006),
		},
	})

	mem := stateinfra.NewMemoryStore()
	txn := mem.BeginTxn("test")
	state.StageWrite(txn, algotrade.StatePendingOrders, algotrade.PendingOrdersState{
		Items: []algotrade.PendingOrder{
			{
				OrderID:     "market_exec-1",
				IntentID:    "intent:decision:USDJPY:buy:2026-04-01T00:04:00Z",
				ExecutionID: "exec:intent:decision:USDJPY:buy:2026-04-01T00:04:00Z:1",
				Symbol:      "USDJPY",
				Intent: algotrade.TradeIntent{
					IntentID:        "intent:decision:USDJPY:buy:2026-04-01T00:04:00Z",
					Symbol:          "USDJPY",
					Action:          algotrade.OrderActionBuy,
					Reason:          "entry_allowed",
					PositionSize:    0.7,
					InitialStopLoss: marketdata.NewPriceFromRaw(20),
				},
				Execution: algotrade.ExecutionResult{
					ExecutionID:   "exec:intent:decision:USDJPY:buy:2026-04-01T00:04:00Z:1",
					IntentID:      "intent:decision:USDJPY:buy:2026-04-01T00:04:00Z",
					OrderID:       "market_exec-1",
					Status:        algotrade.ExecutionStatusSubmitted,
					Action:        algotrade.OrderActionBuy,
					RequestedSize: 0.7,
					Reason:        "submitted",
					RequestedAt:   marketdata.MustParseUTCTime("2026-04-01T00:04:00Z"),
					UpdatedAt:     marketdata.MustParseUTCTime("2026-04-01T00:04:00Z"),
				},
				SourceID: "buy:entry_allowed:0.70000000:20",
			},
		},
	})

	if err := confirmNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run execution_confirm_fill: %v", err)
	}

	result := artifact.MustGet(artifacts.View(), executionFillResultOutputKey("fill_confirm"))
	if !result.Filled {
		t.Fatalf("expected filled=true, got %#v", result)
	}
	execution := artifact.MustGet(artifacts.View(), executionResultOutputKey("fill_confirm"))
	if execution.Status != algotrade.ExecutionStatusFilled {
		t.Fatalf("expected filled execution result, got %#v", execution)
	}
	if execution.ExecutionID != "exec:intent:decision:USDJPY:buy:2026-04-01T00:04:00Z:1" || execution.IntentID == "" {
		t.Fatalf("expected execution/intention propagation, got %#v", execution)
	}
	if execution.FilledSize != 0.7 {
		t.Fatalf("expected filled size propagation, got %#v", execution)
	}
	pending := state.MustGet(txn, algotrade.StatePendingOrders)
	if len(pending.Items) != 0 {
		t.Fatalf("expected pending_orders empty, got %d", len(pending.Items))
	}
	open := state.MustGet(txn, algotrade.StateOpenPositions)
	if len(open.Items) != 1 {
		t.Fatalf("expected one open position, got %d", len(open.Items))
	}
	if !open.Items[0].Snapshot.HasPosition || open.Items[0].Snapshot.EntryPrice.Raw() != 1006 {
		t.Fatalf("expected open long position with entry=1006, got %#v", open.Items[0].Snapshot)
	}
}

func TestExecutionConfirmFillRerunDoesNotDuplicateOpenPosition(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	confirmNode, err := registry.Build(spec.NodeSpec{
		ID:   "fill_confirm",
		Kind: "execution_confirm_fill",
		Config: map[string]any{
			"order_request_node_id": "market_exec",
		},
	})
	if err != nil {
		t.Fatalf("build execution_confirm_fill: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, executionMarketOrderRequestOutputKey("market_exec"), algotrade.MarketOrderRequest{
		Submitted: true,
		OrderID:   "market_exec-1",
		Action:    algotrade.OrderActionBuy,
		Size:      0.7,
		Reason:    "submitted",
		SourceID:  "intent:intent:decision:USDJPY:buy:2026-04-01T00:04:00Z",
	})
	artifact.Set(writer, usecase.InputKeyMarketOHLCVBars, []marketdata.OHLCV{
		{
			Open:  marketdata.NewPriceFromRaw(1000),
			High:  marketdata.NewPriceFromRaw(1008),
			Low:   marketdata.NewPriceFromRaw(998),
			Close: marketdata.NewPriceFromRaw(1006),
		},
	})

	mem := stateinfra.NewMemoryStore()
	txn := mem.BeginTxn("test")
	state.StageWrite(txn, algotrade.StatePendingOrders, algotrade.PendingOrdersState{
		Items: []algotrade.PendingOrder{
			{
				OrderID:     "market_exec-1",
				IntentID:    "intent:decision:USDJPY:buy:2026-04-01T00:04:00Z",
				ExecutionID: "exec:intent:decision:USDJPY:buy:2026-04-01T00:04:00Z:1",
				Symbol:      "USDJPY",
				Intent: algotrade.TradeIntent{
					IntentID:        "intent:decision:USDJPY:buy:2026-04-01T00:04:00Z",
					Symbol:          "USDJPY",
					Action:          algotrade.OrderActionBuy,
					Reason:          "entry_allowed",
					PositionSize:    0.7,
					InitialStopLoss: marketdata.NewPriceFromRaw(20),
				},
				Execution: algotrade.ExecutionResult{
					ExecutionID:   "exec:intent:decision:USDJPY:buy:2026-04-01T00:04:00Z:1",
					IntentID:      "intent:decision:USDJPY:buy:2026-04-01T00:04:00Z",
					OrderID:       "market_exec-1",
					Status:        algotrade.ExecutionStatusSubmitted,
					Action:        algotrade.OrderActionBuy,
					RequestedSize: 0.7,
					Reason:        "submitted",
					RequestedAt:   marketdata.MustParseUTCTime("2026-04-01T00:04:00Z"),
					UpdatedAt:     marketdata.MustParseUTCTime("2026-04-01T00:04:00Z"),
				},
				SourceID: "intent:intent:decision:USDJPY:buy:2026-04-01T00:04:00Z",
			},
		},
	})

	if err := confirmNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("first run execution_confirm_fill: %v", err)
	}
	if err := confirmNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("second run execution_confirm_fill: %v", err)
	}

	open := state.MustGet(txn, algotrade.StateOpenPositions)
	if len(open.Items) != 1 {
		t.Fatalf("expected rerun not to duplicate open position, got %#v", open)
	}
	result := artifact.MustGet(artifacts.View(), executionResultOutputKey("fill_confirm"))
	if result.Reason != "pending_order_not_found" {
		t.Fatalf("expected rerun to be absorbed as missing pending order, got %#v", result)
	}
}

func TestExecutionSubmitMarketOrderFactoryBuildRequiresDecisionNodeID(t *testing.T) {
	t.Parallel()

	factory := &ExecutionSubmitMarketOrderFactory{}
	_, err := factory.Build(spec.NodeSpec{
		ID:     "market_exec",
		Kind:   "execution_submit_market_order",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing decision_node_id")
	}
}

func TestExecutionSubmitMarketOrderRunStagesPendingOrder(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	submitNode, err := registry.Build(spec.NodeSpec{
		ID:   "market_exec",
		Kind: "execution_submit_market_order",
		Config: map[string]any{
			"decision_node_id": "decision",
		},
	})
	if err != nil {
		t.Fatalf("build execution_submit_market_order: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, usecase.InputKeySymbol, "USDJPY")
	artifact.Set(writer, tradeIntentOutputKey("decision"), algotrade.TradeIntent{
		IntentID:        "intent:decision:USDJPY:buy:2026-04-01T00:04:00Z",
		Symbol:          "USDJPY",
		Action:          algotrade.OrderActionBuy,
		Reason:          "entry_allowed",
		PositionSize:    1.25,
		InitialStopLoss: marketdata.NewPriceFromRaw(25),
	})

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	if err := submitNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run execution_submit_market_order: %v", err)
	}

	result := artifact.MustGet(artifacts.View(), executionResultOutputKey("market_exec"))
	if result.Status != algotrade.ExecutionStatusSubmitted {
		t.Fatalf("expected submitted execution result, got %#v", result)
	}
	if result.IntentID != "intent:decision:USDJPY:buy:2026-04-01T00:04:00Z" {
		t.Fatalf("expected intent id propagation into execution result, got %#v", result)
	}
	if result.ExecutionID == "" || result.RequestedSize != 1.25 {
		t.Fatalf("expected execution identity and requested size, got %#v", result)
	}

	request := artifact.MustGet(artifacts.View(), executionMarketOrderRequestOutputKey("market_exec"))
	if !request.Submitted {
		t.Fatalf("expected submitted=true, got %#v", request)
	}
	if request.OrderID == "" {
		t.Fatalf("expected non-empty order id, got %#v", request)
	}

	pending, ok := state.Get(txn, algotrade.StatePendingOrders)
	if !ok {
		t.Fatal("expected pending_orders state staged")
	}
	if len(pending.Items) != 1 {
		t.Fatalf("expected one pending order, got %d", len(pending.Items))
	}
	if pending.Items[0].Symbol != "USDJPY" {
		t.Fatalf("expected pending symbol USDJPY, got %q", pending.Items[0].Symbol)
	}
	if pending.Items[0].IntentID != "intent:decision:USDJPY:buy:2026-04-01T00:04:00Z" {
		t.Fatalf("expected pending intent id propagation, got %#v", pending.Items[0])
	}
	if pending.Items[0].Intent.Action != algotrade.OrderActionBuy || pending.Items[0].Execution.Status != algotrade.ExecutionStatusSubmitted {
		t.Fatalf("expected pending order to keep trade intent and execution payloads, got %#v", pending.Items[0])
	}
	if pending.Items[0].SourceID != "intent:intent:decision:USDJPY:buy:2026-04-01T00:04:00Z" {
		t.Fatalf("expected source id to be intent-derived, got %#v", pending.Items[0])
	}
}

func TestExecutionSubmitMarketOrderHoldProducesRequestedExecutionResult(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	submitNode, err := registry.Build(spec.NodeSpec{
		ID:   "market_exec",
		Kind: "execution_submit_market_order",
		Config: map[string]any{
			"decision_node_id": "decision",
		},
	})
	if err != nil {
		t.Fatalf("build execution_submit_market_order: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, tradeIntentOutputKey("decision"), algotrade.TradeIntent{
		IntentID:  "intent:decision:USDJPY:hold:2026-04-01T00:04:00Z",
		Symbol:    "USDJPY",
		Action:    algotrade.OrderActionHold,
		Reason:    "spread_limit",
		CreatedAt: marketdata.MustParseUTCTime("2026-04-01T00:04:00Z"),
	})

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	if err := submitNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run execution_submit_market_order: %v", err)
	}

	result := artifact.MustGet(artifacts.View(), executionResultOutputKey("market_exec"))
	if result.Status != algotrade.ExecutionStatusRequested || result.Reason != "spread_limit" {
		t.Fatalf("expected requested execution result with reject reason, got %#v", result)
	}
	if result.OrderID != "" || result.RequestedSize != 0 {
		t.Fatalf("expected no order submission for hold intent, got %#v", result)
	}

	pending, ok := state.Get(txn, algotrade.StatePendingOrders)
	if ok && len(pending.Items) > 0 {
		t.Fatalf("expected no pending orders for hold intent, got %#v", pending)
	}
}

func TestExecutionSubmitMarketOrderRerunReusesPendingOrderByIntentID(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	submitNode, err := registry.Build(spec.NodeSpec{
		ID:   "market_exec",
		Kind: "execution_submit_market_order",
		Config: map[string]any{
			"decision_node_id": "decision",
		},
	})
	if err != nil {
		t.Fatalf("build execution_submit_market_order: %v", err)
	}

	intent := algotrade.TradeIntent{
		IntentID:        "intent:decision:USDJPY:buy:2026-04-01T00:04:00Z",
		Symbol:          "USDJPY",
		Action:          algotrade.OrderActionBuy,
		Reason:          "entry_allowed",
		PositionSize:    1.25,
		InitialStopLoss: marketdata.NewPriceFromRaw(25),
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, usecase.InputKeySymbol, "USDJPY")
	artifact.Set(writer, tradeIntentOutputKey("decision"), intent)

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	if err := submitNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("first run execution_submit_market_order: %v", err)
	}

	firstRequest := artifact.MustGet(artifacts.View(), executionMarketOrderRequestOutputKey("market_exec"))
	firstPending := state.MustGet(txn, algotrade.StatePendingOrders)
	if len(firstPending.Items) != 1 {
		t.Fatalf("expected one pending order after first run, got %#v", firstPending)
	}

	if err := submitNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("second run execution_submit_market_order: %v", err)
	}

	secondRequest := artifact.MustGet(artifacts.View(), executionMarketOrderRequestOutputKey("market_exec"))
	secondResult := artifact.MustGet(artifacts.View(), executionResultOutputKey("market_exec"))
	secondPending := state.MustGet(txn, algotrade.StatePendingOrders)

	if len(secondPending.Items) != 1 {
		t.Fatalf("expected pending order dedup on rerun, got %#v", secondPending)
	}
	if secondRequest.OrderID != firstRequest.OrderID {
		t.Fatalf("expected rerun to reuse order id, first=%q second=%q", firstRequest.OrderID, secondRequest.OrderID)
	}
	if secondResult.Reason != "pending_order_exists" || secondResult.OrderID != firstRequest.OrderID {
		t.Fatalf("expected rerun execution result to indicate dedup reuse, got %#v", secondResult)
	}
	if secondPending.Items[0].SourceID != "intent:"+intent.IntentID {
		t.Fatalf("expected intent-derived source id to remain stable, got %#v", secondPending.Items[0])
	}
}

func TestExecutionConfirmFillWithoutSubmittedOrderProducesAcceptedExecutionResult(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	confirmNode, err := registry.Build(spec.NodeSpec{
		ID:   "fill_confirm",
		Kind: "execution_confirm_fill",
		Config: map[string]any{
			"order_request_node_id": "market_exec",
		},
	})
	if err != nil {
		t.Fatalf("build execution_confirm_fill: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, executionMarketOrderRequestOutputKey("market_exec"), algotrade.MarketOrderRequest{
		Submitted: false,
		Action:    algotrade.OrderActionHold,
		Reason:    "spread_limit",
	})

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	if err := confirmNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run execution_confirm_fill: %v", err)
	}

	execution := artifact.MustGet(artifacts.View(), executionResultOutputKey("fill_confirm"))
	if execution.Status != algotrade.ExecutionStatusAccepted || execution.Reason != "spread_limit" {
		t.Fatalf("expected accepted execution result with reason, got %#v", execution)
	}
	fill := artifact.MustGet(artifacts.View(), executionFillResultOutputKey("fill_confirm"))
	if fill.Filled || fill.Reason != "spread_limit" {
		t.Fatalf("expected non-filled compatibility result, got %#v", fill)
	}
}

func TestExecutionConfirmFillRejectedRequestProducesRejectedExecutionResult(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	confirmNode, err := registry.Build(spec.NodeSpec{
		ID:   "fill_confirm",
		Kind: "execution_confirm_fill",
		Config: map[string]any{
			"order_request_node_id": "market_exec",
		},
	})
	if err != nil {
		t.Fatalf("build execution_confirm_fill: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, executionMarketOrderRequestOutputKey("market_exec"), algotrade.MarketOrderRequest{
		Submitted: false,
		Action:    algotrade.OrderActionBuy,
		Reason:    "broker_rejected",
	})

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	if err := confirmNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run execution_confirm_fill: %v", err)
	}

	execution := artifact.MustGet(artifacts.View(), executionResultOutputKey("fill_confirm"))
	if execution.Status != algotrade.ExecutionStatusRejected || execution.Reason != "broker_rejected" {
		t.Fatalf("expected rejected execution result, got %#v", execution)
	}
	fill := artifact.MustGet(artifacts.View(), executionFillResultOutputKey("fill_confirm"))
	if fill.Filled || fill.Reason != "broker_rejected" {
		t.Fatalf("expected rejected compatibility fill result, got %#v", fill)
	}
}

func TestExecutionConfirmFillCancelledRequestProducesCancelledExecutionResult(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	confirmNode, err := registry.Build(spec.NodeSpec{
		ID:   "fill_confirm",
		Kind: "execution_confirm_fill",
		Config: map[string]any{
			"order_request_node_id": "market_exec",
		},
	})
	if err != nil {
		t.Fatalf("build execution_confirm_fill: %v", err)
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, executionMarketOrderRequestOutputKey("market_exec"), algotrade.MarketOrderRequest{
		Submitted: false,
		Action:    algotrade.OrderActionBuy,
		Reason:    "cancelled",
	})

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	if err := confirmNode.Run(context.Background(), artifacts.View(), writer, txn); err != nil {
		t.Fatalf("run execution_confirm_fill: %v", err)
	}

	execution := artifact.MustGet(artifacts.View(), executionResultOutputKey("fill_confirm"))
	if execution.Status != algotrade.ExecutionStatusCancelled || execution.Reason != "cancelled" {
		t.Fatalf("expected cancelled execution result, got %#v", execution)
	}
	fill := artifact.MustGet(artifacts.View(), executionFillResultOutputKey("fill_confirm"))
	if fill.Filled || fill.Reason != "cancelled" {
		t.Fatalf("expected cancelled compatibility fill result, got %#v", fill)
	}
}
