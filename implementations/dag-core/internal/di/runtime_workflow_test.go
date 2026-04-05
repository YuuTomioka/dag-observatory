package di

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	"dag-observatory/dag-core/internal/application/dagruntime/spec/factory"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/application/marketdata/repository"
	"dag-observatory/dag-core/internal/domain/algotrade"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/engine"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/policy"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/dagruntime/workflow"
	"dag-observatory/dag-core/internal/domain/marketdata"
	artifactinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
)

type fakeCompileUnitOfWork struct{}

func (u fakeCompileUnitOfWork) Do(ctx context.Context, fn func(repos repository.Repositories) error) error {
	return nil
}

func (u fakeCompileUnitOfWork) DoReadOnly(ctx context.Context, fn func(repos repository.Repositories) error) error {
	return nil
}

func TestCompileDefaultWorkflowFromYAML(t *testing.T) {
	t.Parallel()

	compiled, err := compileDefaultWorkflow(Config{WorkflowSpecPath: "workflows/default.yaml"}, nil)
	if err != nil {
		t.Fatalf("compile default workflow: %v", err)
	}
	if compiled.Name != "dagruntime.default" {
		t.Fatalf("expected workflow name dagruntime.default, got %q", compiled.Name)
	}
	if len(compiled.Nodes) != 1 {
		t.Fatalf("expected one node, got %d", len(compiled.Nodes))
	}
	if len(compiled.Inputs) != 2 {
		t.Fatalf("expected two inputs, got %d", len(compiled.Inputs))
	}
}

func TestCompileDefaultWorkflowUsesFallbackPath(t *testing.T) {
	t.Parallel()

	compiled, err := compileDefaultWorkflow(Config{}, nil)
	if err != nil {
		t.Fatalf("compile default workflow by fallback path: %v", err)
	}
	if compiled.Name != "dagruntime.default" {
		t.Fatalf("expected workflow name dagruntime.default, got %q", compiled.Name)
	}
}

func TestCompileSMACrossWorkflowFromYAML(t *testing.T) {
	t.Parallel()

	compiled, err := compileWorkflowFromSpecPath("workflows/sma_cross_signal.yaml", nil)
	if err != nil {
		t.Fatalf("compile sma_cross_signal workflow: %v", err)
	}
	if compiled.Name != "dagruntime.sma_cross_signal" {
		t.Fatalf("expected workflow name dagruntime.sma_cross_signal, got %q", compiled.Name)
	}
	if len(compiled.Nodes) != 4 {
		t.Fatalf("expected four nodes, got %d", len(compiled.Nodes))
	}
	if len(compiled.Inputs) != 2 {
		t.Fatalf("expected two inputs, got %d", len(compiled.Inputs))
	}
}

func TestCompileMarketdataBackfillWorkflowFromYAML(t *testing.T) {
	t.Parallel()

	compiled, err := compileWorkflowFromSpecPath("workflows/marketdata_timeframe_bar_backfill.yaml", &MarketDataContainer{
		UnitOfWork: fakeCompileUnitOfWork{},
	})
	if err != nil {
		t.Fatalf("compile marketdata backfill workflow: %v", err)
	}
	if compiled.Name != "dagruntime.marketdata_timeframe_bar_backfill" {
		t.Fatalf("expected workflow name dagruntime.marketdata_timeframe_bar_backfill, got %q", compiled.Name)
	}
	if len(compiled.Nodes) != 5 {
		t.Fatalf("expected five nodes, got %d", len(compiled.Nodes))
	}
	if len(compiled.Inputs) != 4 {
		t.Fatalf("expected four inputs, got %d", len(compiled.Inputs))
	}
}

func TestCompileBreakoutLongV0WorkflowFromYAML(t *testing.T) {
	t.Parallel()

	compiled, err := compileWorkflowFromSpecPath("workflows/breakout_long_v0.yaml", nil)
	if err != nil {
		t.Fatalf("compile breakout_long_v0 workflow: %v", err)
	}
	if compiled.Name != "dagruntime.breakout_long_v0" {
		t.Fatalf("expected workflow name dagruntime.breakout_long_v0, got %q", compiled.Name)
	}
	if len(compiled.Nodes) != 10 {
		t.Fatalf("expected ten nodes, got %d", len(compiled.Nodes))
	}
	if len(compiled.Inputs) != 4 {
		t.Fatalf("expected four inputs, got %d", len(compiled.Inputs))
	}
}

func TestCompileBreakoutExecutionPositionMinimalWorkflowFromYAML(t *testing.T) {
	t.Parallel()

	compiled, err := compileWorkflowFromSpecPath("workflows/breakout_long_execution_position_minimal.yaml", nil)
	if err != nil {
		t.Fatalf("compile breakout_long_execution_position_minimal workflow: %v", err)
	}
	if compiled.Name != "dagruntime.breakout_long_execution_position_minimal" {
		t.Fatalf("expected workflow name dagruntime.breakout_long_execution_position_minimal, got %q", compiled.Name)
	}
	if len(compiled.Nodes) != 9 {
		t.Fatalf("expected nine nodes, got %d", len(compiled.Nodes))
	}
	if len(compiled.Inputs) != 4 {
		t.Fatalf("expected four inputs, got %d", len(compiled.Inputs))
	}
}

func TestCompileBreakoutLongV1ExtendedWorkflowFromYAML(t *testing.T) {
	t.Parallel()

	compiled, err := compileWorkflowFromSpecPath("workflows/breakout_long_v1_extended.yaml", nil)
	if err != nil {
		t.Fatalf("compile breakout_long_v1_extended workflow: %v", err)
	}
	if compiled.Name != "dagruntime.breakout_long_v1_extended" {
		t.Fatalf("expected workflow name dagruntime.breakout_long_v1_extended, got %q", compiled.Name)
	}
	if len(compiled.Nodes) != 17 {
		t.Fatalf("expected seventeen nodes, got %d", len(compiled.Nodes))
	}
	if len(compiled.Inputs) != 4 {
		t.Fatalf("expected four inputs, got %d", len(compiled.Inputs))
	}
}

func TestCompileBreakoutLongResultReflectionMinimalWorkflowFromYAML(t *testing.T) {
	t.Parallel()

	compiled, err := compileWorkflowFromSpecPath("workflows/breakout_long_result_reflection_minimal.yaml", nil)
	if err != nil {
		t.Fatalf("compile breakout_long_result_reflection_minimal workflow: %v", err)
	}
	if compiled.Name != "dagruntime.breakout_long_result_reflection_minimal" {
		t.Fatalf("expected workflow name dagruntime.breakout_long_result_reflection_minimal, got %q", compiled.Name)
	}
	if len(compiled.Nodes) != 9 {
		t.Fatalf("expected nine nodes, got %d", len(compiled.Nodes))
	}
	if len(compiled.Inputs) != 1 {
		t.Fatalf("expected one input, got %d", len(compiled.Inputs))
	}
}

func TestRunBreakoutLongV0WorkflowE2E(t *testing.T) {
	t.Parallel()

	compiled, err := compileWorkflowFromSpecPath("workflows/breakout_long_v0.yaml", nil)
	if err != nil {
		t.Fatalf("compile breakout_long_v0 workflow: %v", err)
	}

	artifactStore := artifactinfra.NewMemoryStore()
	runner := &engine.Runner{
		ArtifactStore: artifactStore,
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{},
	}

	inputs := engine.InputMap{
		usecase.InputKeySymbol: "USDJPY",
		usecase.InputKeyMarketOHLCVBars: []marketdata.OHLCV{
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
		},
		usecase.InputKeyMarketSpreadBps: 1.2,
		usecase.InputKeyAccountBalance:  10000.0,
	}

	event := events.Event{
		EventID:   "breakout-v0-e2e",
		EventTime: time.Now().UTC(),
		Partition: state.Partition("breakout-v0-e2e"),
		Type:      "task.requested",
		Payload:   nil,
	}
	if err := runner.RunCycle(context.Background(), compiled, inputs, event.Partition, event); err != nil {
		t.Fatalf("run breakout_long_v0 workflow: %v", err)
	}

	decisionKey := artifact.Key[any]{
		Name:     "decision.signal_order_decision",
		StableID: "artifact:dagruntime.signal.decision.decision.v1",
	}
	observeKey := artifact.Key[any]{
		Name:     "observe.observability_signal_decision",
		StableID: "artifact:dagruntime.observability.signal_decision.observe.v1",
	}
	if !artifactStore.Has(decisionKey) {
		t.Fatalf("expected decision artifact key %s", decisionKey.String())
	}
	if !artifactStore.Has(observeKey) {
		t.Fatalf("expected observability artifact key %s", observeKey.String())
	}
}

func TestRunBreakoutExecutionPositionMinimalWorkflowE2E(t *testing.T) {
	t.Parallel()

	compiled, err := compileWorkflowFromSpecPath("workflows/breakout_long_execution_position_minimal.yaml", nil)
	if err != nil {
		t.Fatalf("compile breakout_long_execution_position_minimal workflow: %v", err)
	}

	artifactStore := artifactinfra.NewMemoryStore()
	stateStore := stateinfra.NewMemoryStore()
	runner := &engine.Runner{
		ArtifactStore: artifactStore,
		StateStore:    stateStore,
		Policy:        policy.Policy{},
	}

	event := events.Event{
		EventID:   "breakout-minimal-e2e",
		EventTime: time.Now().UTC(),
		Partition: state.Partition("breakout-minimal-e2e"),
		Type:      "task.requested",
		Payload:   nil,
	}
	if err := runner.RunCycle(context.Background(), compiled, breakoutWorkflowInputs(), event.Partition, event); err != nil {
		t.Fatalf("run breakout_long_execution_position_minimal workflow: %v", err)
	}

	marketExecKey := artifact.Key[algotrade.MarketOrderRequest]{
		Name:     "market_exec.market_order_request",
		StableID: "artifact:dagruntime.execution.market_order_request.market_exec.v1",
	}
	positionLoadKey := artifact.Key[algotrade.PositionSnapshot]{
		Name:     "position_load.position_snapshot",
		StableID: "artifact:dagruntime.position.snapshot.position_load.v1",
	}
	marketExec, ok := artifact.Get(artifactStore.View(), marketExecKey)
	if !ok {
		t.Fatalf("expected market execution artifact key %s", marketExecKey.String())
	}
	if !marketExec.Submitted {
		t.Fatalf("expected submitted market order request, got %+v", marketExec)
	}
	position, ok := artifact.Get(artifactStore.View(), positionLoadKey)
	if !ok {
		t.Fatalf("expected position snapshot artifact key %s", positionLoadKey.String())
	}
	if position.HasPosition {
		t.Fatalf("expected no open position before fill confirmation, got %+v", position)
	}

	txn := stateStore.BeginTxn(event.Partition)
	pending, ok := state.Get(txn, algotrade.StatePendingOrders)
	if !ok {
		t.Fatal("expected pending orders state to be written")
	}
	if len(pending.Items) != 1 {
		t.Fatalf("expected one pending order, got %d", len(pending.Items))
	}
}

func TestRunBreakoutSubmitAndFillAcrossSeparateWorkflows(t *testing.T) {
	t.Parallel()

	submitCompiled, err := compileWorkflowFromSpecPath("workflows/breakout_long_execution_position_minimal.yaml", nil)
	if err != nil {
		t.Fatalf("compile submit workflow: %v", err)
	}

	registry, err := factory.NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	fillConfirmNode, err := registry.Build(spec.NodeSpec{
		ID:   "fill_confirm",
		Kind: "execution_confirm_fill",
		Config: map[string]any{
			"order_request_node_id": "market_exec",
		},
	})
	if err != nil {
		t.Fatalf("build fill confirm node: %v", err)
	}
	marketExecKey := artifact.Key[algotrade.MarketOrderRequest]{
		Name:     "market_exec.market_order_request",
		StableID: "artifact:dagruntime.execution.market_order_request.market_exec.v1",
	}
	fillCompiled, err := workflow.Compile(workflow.Workflow{
		Name:   "dagruntime.breakout_long_fill_confirm_minimal",
		Inputs: []artifact.AnyKey{marketExecKey},
		Nodes:  []node.Node{fillConfirmNode},
	})
	if err != nil {
		t.Fatalf("compile fill workflow: %v", err)
	}

	artifactStore := artifactinfra.NewMemoryStore()
	stateStore := stateinfra.NewMemoryStore()
	runner := &engine.Runner{
		ArtifactStore: artifactStore,
		StateStore:    stateStore,
		Policy:        policy.Policy{},
	}

	partition := state.Partition("breakout-split-workflows-e2e")
	submitEvent := events.Event{
		EventID:   "breakout-submit-cycle",
		EventTime: time.Now().UTC(),
		Partition: partition,
		Type:      "task.requested",
		Payload:   nil,
	}
	if err := runner.RunCycle(context.Background(), submitCompiled, breakoutWorkflowInputs(), partition, submitEvent); err != nil {
		t.Fatalf("run submit workflow: %v", err)
	}

	marketReq, ok := artifact.Get(artifactStore.View(), marketExecKey)
	if !ok {
		t.Fatalf("expected submit workflow output artifact key %s", marketExecKey.String())
	}
	if !marketReq.Submitted || marketReq.OrderID == "" {
		t.Fatalf("expected submitted order request from submit workflow, got %+v", marketReq)
	}

	fillEvent := events.Event{
		EventID:   "breakout-fill-cycle",
		EventTime: time.Now().UTC(),
		Partition: partition,
		Type:      "task.completed",
		Payload:   nil,
	}
	if err := runner.RunCycle(context.Background(), fillCompiled, engine.InputMap{marketExecKey: marketReq}, partition, fillEvent); err != nil {
		t.Fatalf("run fill workflow: %v", err)
	}

	fillResultKey := artifact.Key[algotrade.FillResult]{
		Name:     "fill_confirm.fill_result",
		StableID: "artifact:dagruntime.execution.fill_result.fill_confirm.v1",
	}
	fillResult, ok := artifact.Get(artifactStore.View(), fillResultKey)
	if !ok {
		t.Fatalf("expected fill result artifact key %s", fillResultKey.String())
	}
	if !fillResult.Filled {
		t.Fatalf("expected fill result filled=true, got %+v", fillResult)
	}

	txn := stateStore.BeginTxn(partition)
	pending, _ := state.Get(txn, algotrade.StatePendingOrders)
	if len(pending.Items) != 0 {
		t.Fatalf("expected pending orders to be empty after fill, got %d", len(pending.Items))
	}
	openPositions, ok := state.Get(txn, algotrade.StateOpenPositions)
	if !ok || len(openPositions.Items) != 1 {
		t.Fatalf("expected one open position after fill, got %+v", openPositions)
	}
}

func TestRunBreakoutResultReflectionMinimalWorkflowE2E(t *testing.T) {
	t.Parallel()

	compiled, err := compileWorkflowFromSpecPath("workflows/breakout_long_result_reflection_minimal.yaml", nil)
	if err != nil {
		t.Fatalf("compile breakout_long_result_reflection_minimal workflow: %v", err)
	}

	artifactStore := artifactinfra.NewMemoryStore()
	stateStore := stateinfra.NewMemoryStore()
	runner := &engine.Runner{
		ArtifactStore: artifactStore,
		StateStore:    stateStore,
		Policy:        policy.Policy{},
	}

	partition := state.Partition("breakout-result-reflection-e2e")
	txn := stateStore.BeginTxn(partition)
	state.StageWrite(txn, algotrade.StateOpenPositions, algotrade.OpenPositionsState{
		Items: []algotrade.OpenPosition{
			{
				PositionID:      "position-1",
				IntentID:        "intent-1",
				Symbol:          "USDJPY",
				Side:            algotrade.PositionSideLong,
				Size:            1,
				EntryPrice:      marketdata.NewPriceFromRaw(1000),
				EntryTime:       marketdata.MustParseUTCTime("2026-04-01T00:00:00Z"),
				StrategyID:      "breakout",
				WorkflowName:    "dagruntime.breakout_long_execution_position_minimal",
				WorkflowVersion: "v1",
				ParameterSetID:  "p1",
			},
		},
	})
	if err := txn.Commit(); err != nil {
		t.Fatalf("commit seeded state: %v", err)
	}

	event := events.Event{
		EventID:   "breakout-result-reflection",
		EventTime: time.Now().UTC(),
		Partition: partition,
		Type:      "task.completed",
		Payload:   nil,
	}
	inputs := engine.InputMap{
		usecase.InputKeyMarketOHLCVBars: []marketdata.OHLCV{
			{
				Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:04:00Z"),
				Closetime: marketdata.MustParseUTCTime("2026-04-01T00:05:00Z"),
				Open:      marketdata.NewPriceFromRaw(1000),
				High:      marketdata.NewPriceFromRaw(1002),
				Low:       marketdata.NewPriceFromRaw(998),
				Close:     marketdata.NewPriceFromRaw(993),
			},
		},
	}
	if err := runner.RunCycle(context.Background(), compiled, inputs, partition, event); err != nil {
		t.Fatalf("run breakout_long_result_reflection_minimal workflow: %v", err)
	}

	readTxn := stateStore.BeginTxn(partition)
	closedTrades, ok := state.Get(readTxn, algotrade.StateClosedTrades)
	if !ok || len(closedTrades.Items) != 1 {
		t.Fatalf("expected one closed trade, got %+v", closedTrades)
	}
	if closedTrades.Items[0].ExitReason != "stop_loss_hit" {
		t.Fatalf("expected stop_loss_hit exit reason, got %+v", closedTrades.Items[0])
	}

	dailyPnL, ok := state.Get(readTxn, algotrade.StateDailyPnL)
	if !ok || dailyPnL.RealizedPnL != -7 {
		t.Fatalf("expected realized pnl -7, got %+v", dailyPnL)
	}

	openPositions, _ := state.Get(readTxn, algotrade.StateOpenPositions)
	if len(openPositions.Items) != 0 {
		t.Fatalf("expected open position to be removed, got %+v", openPositions)
	}

	summary, ok := state.Get(readTxn, algotrade.StateStrategySummary)
	if !ok || summary.Summary.TradeCount != 1 || summary.Summary.LossCount != 1 {
		t.Fatalf("expected summary to reflect one losing trade, got %+v", summary)
	}
}

func TestRunBreakoutSubmitWorkflowIsIdempotentForPendingOrders(t *testing.T) {
	t.Parallel()

	compiled, err := compileWorkflowFromSpecPath("workflows/breakout_long_execution_position_minimal.yaml", nil)
	if err != nil {
		t.Fatalf("compile breakout_long_execution_position_minimal workflow: %v", err)
	}

	artifactStore := artifactinfra.NewMemoryStore()
	stateStore := stateinfra.NewMemoryStore()
	runner := &engine.Runner{
		ArtifactStore: artifactStore,
		StateStore:    stateStore,
		Policy:        policy.Policy{},
	}

	partition := state.Partition("breakout-submit-idempotent")
	run := func(eventID string) {
		t.Helper()
		event := events.Event{
			EventID:   eventID,
			EventTime: time.Now().UTC(),
			Partition: partition,
			Type:      "task.requested",
			Payload:   nil,
		}
		if err := runner.RunCycle(context.Background(), compiled, breakoutWorkflowInputs(), partition, event); err != nil {
			t.Fatalf("run submit workflow (%s): %v", eventID, err)
		}
	}

	run("submit-cycle-1")
	firstReqKey := artifact.Key[algotrade.MarketOrderRequest]{
		Name:     "market_exec.market_order_request",
		StableID: "artifact:dagruntime.execution.market_order_request.market_exec.v1",
	}
	firstReq, ok := artifact.Get(artifactStore.View(), firstReqKey)
	if !ok || !firstReq.Submitted || firstReq.OrderID == "" {
		t.Fatalf("expected submitted request after first cycle, got %+v", firstReq)
	}

	run("submit-cycle-2")
	secondReq, ok := artifact.Get(artifactStore.View(), firstReqKey)
	if !ok || !secondReq.Submitted {
		t.Fatalf("expected submitted request after second cycle, got %+v", secondReq)
	}
	if secondReq.OrderID != firstReq.OrderID {
		t.Fatalf("expected stable order id across retries, first=%q second=%q", firstReq.OrderID, secondReq.OrderID)
	}
	if secondReq.Reason != "pending_order_exists" {
		t.Fatalf("expected pending_order_exists on second cycle, got %q", secondReq.Reason)
	}

	txn := stateStore.BeginTxn(partition)
	pending, ok := state.Get(txn, algotrade.StatePendingOrders)
	if !ok {
		t.Fatal("expected pending orders state")
	}
	if len(pending.Items) != 1 {
		t.Fatalf("expected one pending order after duplicate submit cycles, got %d", len(pending.Items))
	}
}

func breakoutWorkflowInputs() engine.InputMap {
	return engine.InputMap{
		usecase.InputKeySymbol: "USDJPY",
		usecase.InputKeyMarketOHLCVBars: []marketdata.OHLCV{
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
		},
		usecase.InputKeyMarketSpreadBps: 1.2,
		usecase.InputKeyAccountBalance:  10000.0,
	}
}

func TestCompileWorkflowRejectsSubmitAndConfirmInSameWorkflow(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "breakout_submit_confirm_same_workflow.yaml")
	raw := []byte(`name: dagruntime.breakout_submit_confirm_same_workflow
version: v1
inputs:
  - market.symbol
  - market.ohlcv_bars
  - market.spread_bps
  - account.balance
nodes:
  - id: atr
    kind: feature_atr
    config:
      window: 14
  - id: range_high
    kind: feature_range_high
    config:
      window: 20
  - id: breakout
    kind: signal_breakout_long
    config:
      range_node_id: range_high
  - id: session_filter
    kind: filter_session
    config:
      signal_node_id: breakout
      allowed_sessions:
        - tokyo
        - london
  - id: spread_filter
    kind: filter_spread
    config:
      allowed_node_id: session_filter
      max_spread_bps: 5
  - id: sizing
    kind: risk_position_sizing
    config:
      allowed_node_id: spread_filter
      atr_node_id: atr
      risk_rate: 0.005
      stop_atr_multiple: 1.5
  - id: decision
    kind: signal_decision_mapper
    config:
      allowed_node_id: spread_filter
      sizing_node_id: sizing
  - id: market_exec
    kind: execution_submit_market_order
    config:
      decision_node_id: decision
  - id: fill_confirm
    kind: execution_confirm_fill
    config:
      order_request_node_id: market_exec
`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write invalid workflow yaml: %v", err)
	}

	_, err := compileWorkflowFromSpecPath(path, nil)
	if err == nil {
		t.Fatal("expected compile error for duplicate writer")
	}
	if !strings.Contains(err.Error(), "duplicate_writer") {
		t.Fatalf("expected duplicate_writer error, got %v", err)
	}
}

func TestCompileWorkflowFromYAMLDetectsConfigError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "invalid_breakout.yaml")
	raw := []byte(`name: dagruntime.breakout_invalid
version: v1
inputs:
  - market.symbol
  - market.ohlcv_bars
nodes:
  - id: atr
    kind: feature_atr
    config: {}
`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write invalid workflow yaml: %v", err)
	}

	if _, err := compileWorkflowFromSpecPath(path, nil); err == nil {
		t.Fatal("expected compile error for invalid feature_atr config")
	}
}
