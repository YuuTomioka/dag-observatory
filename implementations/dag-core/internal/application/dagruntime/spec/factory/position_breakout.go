package factory

import (
	"context"
	"fmt"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/algotrade"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

type PositionSnapshotLoadFactory struct{}

func (f *PositionSnapshotLoadFactory) Kind() string { return "position_snapshot_load" }

func (f *PositionSnapshotLoadFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, nil); err != nil {
		return nil, fmt.Errorf("position_snapshot_load node %q: %w", nodeSpec.ID, err)
	}
	return &positionSnapshotLoadNode{id: nodeSpec.ID}, nil
}

type PositionBreakevenFactory struct{}

func (f *PositionBreakevenFactory) Kind() string { return "position_breakeven" }

func (f *PositionBreakevenFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"position_node_id"}); err != nil {
		return nil, fmt.Errorf("position_breakeven node %q: %w", nodeSpec.ID, err)
	}
	positionNodeID, err := requiredString(nodeSpec.Config, "position_node_id")
	if err != nil {
		return nil, fmt.Errorf("position_breakeven node %q: %w", nodeSpec.ID, err)
	}
	return &positionBreakevenNode{
		id:             nodeSpec.ID,
		positionNodeID: positionNodeID,
	}, nil
}

type PositionTrailingStopFactory struct{}

func (f *PositionTrailingStopFactory) Kind() string { return "position_trailing_stop" }

func (f *PositionTrailingStopFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"position_node_id"}); err != nil {
		return nil, fmt.Errorf("position_trailing_stop node %q: %w", nodeSpec.ID, err)
	}
	positionNodeID, err := requiredString(nodeSpec.Config, "position_node_id")
	if err != nil {
		return nil, fmt.Errorf("position_trailing_stop node %q: %w", nodeSpec.ID, err)
	}
	return &positionTrailingStopNode{
		id:             nodeSpec.ID,
		positionNodeID: positionNodeID,
	}, nil
}

type PositionTimeoutExitFactory struct{}

func (f *PositionTimeoutExitFactory) Kind() string { return "position_timeout_exit" }

func (f *PositionTimeoutExitFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"position_node_id"}); err != nil {
		return nil, fmt.Errorf("position_timeout_exit node %q: %w", nodeSpec.ID, err)
	}
	positionNodeID, err := requiredString(nodeSpec.Config, "position_node_id")
	if err != nil {
		return nil, fmt.Errorf("position_timeout_exit node %q: %w", nodeSpec.ID, err)
	}
	return &positionTimeoutExitNode{
		id:             nodeSpec.ID,
		positionNodeID: positionNodeID,
	}, nil
}

type PositionTrackerUpdateFactory struct{}

func (f *PositionTrackerUpdateFactory) Kind() string { return "position_tracker_update" }

func (f *PositionTrackerUpdateFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"execution_node_id"}); err != nil {
		return nil, fmt.Errorf("position_tracker_update node %q: %w", nodeSpec.ID, err)
	}
	executionNodeID, err := requiredString(nodeSpec.Config, "execution_node_id")
	if err != nil {
		return nil, fmt.Errorf("position_tracker_update node %q: %w", nodeSpec.ID, err)
	}
	return &positionTrackerUpdateNode{
		id:              nodeSpec.ID,
		executionNodeID: executionNodeID,
	}, nil
}

type PositionCloseToClosedTradeFactory struct{}

func (f *PositionCloseToClosedTradeFactory) Kind() string { return "position_close_to_closed_trade" }

func (f *PositionCloseToClosedTradeFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"position_node_id", "exit_node_id"}); err != nil {
		return nil, fmt.Errorf("position_close_to_closed_trade node %q: %w", nodeSpec.ID, err)
	}
	positionNodeID, err := requiredString(nodeSpec.Config, "position_node_id")
	if err != nil {
		return nil, fmt.Errorf("position_close_to_closed_trade node %q: %w", nodeSpec.ID, err)
	}
	exitNodeID, err := requiredString(nodeSpec.Config, "exit_node_id")
	if err != nil {
		return nil, fmt.Errorf("position_close_to_closed_trade node %q: %w", nodeSpec.ID, err)
	}
	return &positionCloseToClosedTradeNode{
		id:             nodeSpec.ID,
		positionNodeID: positionNodeID,
		exitNodeID:     exitNodeID,
	}, nil
}

type ClosedTradeStoreFactory struct{}

func (f *ClosedTradeStoreFactory) Kind() string { return "closed_trade_store" }

func (f *ClosedTradeStoreFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"trade_node_id"}); err != nil {
		return nil, fmt.Errorf("closed_trade_store node %q: %w", nodeSpec.ID, err)
	}
	tradeNodeID, err := requiredString(nodeSpec.Config, "trade_node_id")
	if err != nil {
		return nil, fmt.Errorf("closed_trade_store node %q: %w", nodeSpec.ID, err)
	}
	return &closedTradeStoreNode{
		id:          nodeSpec.ID,
		tradeNodeID: tradeNodeID,
	}, nil
}

type DailyPnLUpdateFactory struct{}

func (f *DailyPnLUpdateFactory) Kind() string { return "daily_pnl_update" }

func (f *DailyPnLUpdateFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"trade_node_id"}); err != nil {
		return nil, fmt.Errorf("daily_pnl_update node %q: %w", nodeSpec.ID, err)
	}
	tradeNodeID, err := requiredString(nodeSpec.Config, "trade_node_id")
	if err != nil {
		return nil, fmt.Errorf("daily_pnl_update node %q: %w", nodeSpec.ID, err)
	}
	return &dailyPnLUpdateNode{
		id:          nodeSpec.ID,
		tradeNodeID: tradeNodeID,
	}, nil
}

type OpenPositionCloseFactory struct{}

func (f *OpenPositionCloseFactory) Kind() string { return "open_position_close" }

func (f *OpenPositionCloseFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"trade_node_id"}); err != nil {
		return nil, fmt.Errorf("open_position_close node %q: %w", nodeSpec.ID, err)
	}
	tradeNodeID, err := requiredString(nodeSpec.Config, "trade_node_id")
	if err != nil {
		return nil, fmt.Errorf("open_position_close node %q: %w", nodeSpec.ID, err)
	}
	return &openPositionCloseNode{
		id:          nodeSpec.ID,
		tradeNodeID: tradeNodeID,
	}, nil
}

type StrategySummaryUpdateFactory struct{}

func (f *StrategySummaryUpdateFactory) Kind() string { return "strategy_summary_update" }

func (f *StrategySummaryUpdateFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"trade_node_id"}); err != nil {
		return nil, fmt.Errorf("strategy_summary_update node %q: %w", nodeSpec.ID, err)
	}
	tradeNodeID, err := requiredString(nodeSpec.Config, "trade_node_id")
	if err != nil {
		return nil, fmt.Errorf("strategy_summary_update node %q: %w", nodeSpec.ID, err)
	}
	return &strategySummaryUpdateNode{
		id:          nodeSpec.ID,
		tradeNodeID: tradeNodeID,
	}, nil
}

type positionSnapshotLoadNode struct {
	id string
}

func (n *positionSnapshotLoadNode) Name() string { return "dagruntime.position_snapshot_load." + n.id }
func (n *positionSnapshotLoadNode) Requires() []artifact.AnyKey {
	return nil
}
func (n *positionSnapshotLoadNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{positionSnapshotOutputKey(n.id)}
}
func (n *positionSnapshotLoadNode) Reads() []state.AnyKey {
	return []state.AnyKey{algotrade.StateOpenPositions}
}
func (n *positionSnapshotLoadNode) Writes() []state.AnyKey { return nil }
func (n *positionSnapshotLoadNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}

func (n *positionSnapshotLoadNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = av
	openPositions, ok := state.Get(txn, algotrade.StateOpenPositions)
	if !ok || len(openPositions.Items) == 0 {
		artifact.Set(aw, positionSnapshotOutputKey(n.id), algotrade.PositionSnapshot{
			HasPosition: false,
			Side:        algotrade.PositionSideFlat,
			Size:        0,
			EntryPrice:  marketdata.Price(0),
		})
		return nil
	}
	artifact.Set(aw, positionSnapshotOutputKey(n.id), openPositions.Items[0].SnapshotView())
	return nil
}

type positionBreakevenNode struct {
	id             string
	positionNodeID string
}

func (n *positionBreakevenNode) Name() string { return "dagruntime.position_breakeven." + n.id }
func (n *positionBreakevenNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		positionSnapshotOutputKey(n.positionNodeID),
		usecase.InputKeyMarketOHLCVBars,
	}
}
func (n *positionBreakevenNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{riskStopLossDistanceOutputKey(n.id)}
}
func (n *positionBreakevenNode) Reads() []state.AnyKey {
	return []state.AnyKey{algotrade.StateOpenPositions}
}
func (n *positionBreakevenNode) Writes() []state.AnyKey {
	return []state.AnyKey{algotrade.StateOpenPositions}
}
func (n *positionBreakevenNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *positionBreakevenNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	position := artifact.MustGet(av, positionSnapshotOutputKey(n.positionNodeID))
	if !position.HasPosition || position.Side != algotrade.PositionSideLong || position.Size <= 0 {
		artifact.Set(aw, riskStopLossDistanceOutputKey(n.id), marketdata.Price(1))
		return nil
	}
	openPositions, _ := state.Get(txn, algotrade.StateOpenPositions)
	state.StageWrite(txn, algotrade.StateOpenPositions, openPositions)
	artifact.Set(aw, riskStopLossDistanceOutputKey(n.id), marketdata.Price(1))
	return nil
}

type positionTrailingStopNode struct {
	id             string
	positionNodeID string
}

func (n *positionTrailingStopNode) Name() string { return "dagruntime.position_trailing_stop." + n.id }
func (n *positionTrailingStopNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		positionSnapshotOutputKey(n.positionNodeID),
		usecase.InputKeyMarketOHLCVBars,
	}
}
func (n *positionTrailingStopNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{riskStopLossDistanceOutputKey(n.id), riskTrailingStopDistanceOutputKey(n.id)}
}
func (n *positionTrailingStopNode) Reads() []state.AnyKey {
	return []state.AnyKey{algotrade.StateOpenPositions}
}
func (n *positionTrailingStopNode) Writes() []state.AnyKey {
	return []state.AnyKey{algotrade.StateOpenPositions}
}
func (n *positionTrailingStopNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *positionTrailingStopNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	position := artifact.MustGet(av, positionSnapshotOutputKey(n.positionNodeID))
	bars := artifact.MustGet(av, usecase.InputKeyMarketOHLCVBars)
	if len(bars) == 0 {
		return fmt.Errorf("position_trailing_stop node %q: market.ohlcv_bars is empty", n.id)
	}
	stopDistance := marketdata.Price(1)
	if position.HasPosition && position.Side == algotrade.PositionSideLong && position.Size > 0 {
		last := bars[len(bars)-1].Close
		if last.Gt(position.EntryPrice) {
			profitDistance := last.Sub(position.EntryPrice).Raw()
			if profitDistance > 1 {
				stopDistance = marketdata.NewPriceFromRaw(profitDistance / 2)
			}
		}
	}
	openPositions, _ := state.Get(txn, algotrade.StateOpenPositions)
	state.StageWrite(txn, algotrade.StateOpenPositions, openPositions)
	artifact.Set(aw, riskStopLossDistanceOutputKey(n.id), stopDistance)
	artifact.Set(aw, riskTrailingStopDistanceOutputKey(n.id), stopDistance)
	return nil
}

type positionTimeoutExitNode struct {
	id             string
	positionNodeID string
}

func (n *positionTimeoutExitNode) Name() string { return "dagruntime.position_timeout_exit." + n.id }
func (n *positionTimeoutExitNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{positionSnapshotOutputKey(n.positionNodeID)}
}
func (n *positionTimeoutExitNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{signalExitDecisionOutputKey(n.id)}
}
func (n *positionTimeoutExitNode) Reads() []state.AnyKey {
	return []state.AnyKey{algotrade.StateOpenPositions}
}
func (n *positionTimeoutExitNode) Writes() []state.AnyKey { return nil }
func (n *positionTimeoutExitNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *positionTimeoutExitNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	position := artifact.MustGet(av, positionSnapshotOutputKey(n.positionNodeID))
	decision := algotrade.ExitDecision{
		ShouldExit: false,
		Reason:     "timeout_not_reached",
	}
	if !position.HasPosition || position.Side == algotrade.PositionSideFlat || position.Size <= 0 {
		decision.Reason = "no_position"
	}
	artifact.Set(aw, signalExitDecisionOutputKey(n.id), decision)
	return nil
}

type positionTrackerUpdateNode struct {
	id              string
	executionNodeID string
}

type positionCloseToClosedTradeNode struct {
	id             string
	positionNodeID string
	exitNodeID     string
}

type closedTradeStoreNode struct {
	id          string
	tradeNodeID string
}

type dailyPnLUpdateNode struct {
	id          string
	tradeNodeID string
}

type openPositionCloseNode struct {
	id          string
	tradeNodeID string
}

type strategySummaryUpdateNode struct {
	id          string
	tradeNodeID string
}

func (n *positionTrackerUpdateNode) Name() string {
	return "dagruntime.position_tracker_update." + n.id
}
func (n *positionTrackerUpdateNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		paperExecutionResultOutputKey(n.executionNodeID),
		usecase.InputKeyMarketOHLCVBars,
	}
}
func (n *positionTrackerUpdateNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{positionSnapshotOutputKey(n.id)}
}
func (n *positionTrackerUpdateNode) Reads() []state.AnyKey  { return nil }
func (n *positionTrackerUpdateNode) Writes() []state.AnyKey { return nil }
func (n *positionTrackerUpdateNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}

func (n *positionTrackerUpdateNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	execution := artifact.MustGet(av, paperExecutionResultOutputKey(n.executionNodeID))
	bars := artifact.MustGet(av, usecase.InputKeyMarketOHLCVBars)
	if len(bars) == 0 {
		return fmt.Errorf("position_tracker_update node %q: market.ohlcv_bars is empty", n.id)
	}
	snapshot := algotrade.PositionSnapshot{
		HasPosition: false,
		Side:        algotrade.PositionSideFlat,
		Size:        0,
		EntryPrice:  marketdata.Price(0),
	}
	if execution.Submitted && execution.Action == algotrade.OrderActionBuy && execution.Size > 0 {
		last := bars[len(bars)-1]
		snapshot = algotrade.PositionSnapshot{
			HasPosition: true,
			Side:        algotrade.PositionSideLong,
			Size:        execution.Size,
			EntryPrice:  last.Close,
		}
	}
	artifact.Set(aw, positionSnapshotOutputKey(n.id), snapshot)
	return nil
}

func (n *positionCloseToClosedTradeNode) Name() string {
	return "dagruntime.position_close_to_closed_trade." + n.id
}
func (n *positionCloseToClosedTradeNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		positionSnapshotOutputKey(n.positionNodeID),
		signalExitDecisionOutputKey(n.exitNodeID),
		usecase.InputKeyMarketOHLCVBars,
	}
}
func (n *positionCloseToClosedTradeNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{closedTradeOutputKey(n.id)}
}
func (n *positionCloseToClosedTradeNode) Reads() []state.AnyKey {
	return []state.AnyKey{algotrade.StateOpenPositions}
}
func (n *positionCloseToClosedTradeNode) Writes() []state.AnyKey { return nil }
func (n *positionCloseToClosedTradeNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *positionCloseToClosedTradeNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	position := artifact.MustGet(av, positionSnapshotOutputKey(n.positionNodeID))
	decision := artifact.MustGet(av, signalExitDecisionOutputKey(n.exitNodeID))
	if !decision.ShouldExit || !position.HasPosition || position.Side == algotrade.PositionSideFlat || position.Size <= 0 {
		artifact.Set(aw, closedTradeOutputKey(n.id), algotrade.ClosedTrade{})
		return nil
	}

	openPositions, ok := state.Get(txn, algotrade.StateOpenPositions)
	if !ok || len(openPositions.Items) == 0 {
		artifact.Set(aw, closedTradeOutputKey(n.id), algotrade.ClosedTrade{})
		return nil
	}
	open := openPositions.Items[0]
	bars := artifact.MustGet(av, usecase.InputKeyMarketOHLCVBars)
	if len(bars) == 0 {
		return fmt.Errorf("position_close_to_closed_trade node %q: market.ohlcv_bars is empty", n.id)
	}
	last := bars[len(bars)-1]
	exitPrice := last.Close
	exitTime := last.Closetime
	if exitTime.IsZero() {
		exitTime = last.Opentime
	}
	if exitTime.IsZero() {
		exitTime = marketdata.NowUTCTime()
	}
	grossPnL := closedTradeGrossPnL(open.Side, open.EntryPrice, exitPrice, open.Size)
	holdingDuration := ""
	if !open.EntryTime.IsZero() && !exitTime.IsZero() {
		holdingDuration = exitTime.Time().Sub(open.EntryTime.Time()).String()
	}
	artifact.Set(aw, closedTradeOutputKey(n.id), algotrade.ClosedTrade{
		TradeID:         closedTradeID(open),
		IntentID:        open.IntentID,
		PositionID:      open.PositionID,
		Symbol:          open.Symbol,
		Side:            open.Side,
		Size:            open.Size,
		EntryTime:       open.EntryTime,
		ExitTime:        exitTime,
		EntryPrice:      open.EntryPrice,
		ExitPrice:       exitPrice,
		GrossPnL:        grossPnL,
		NetPnL:          closedTradeNetPnL(grossPnL),
		ExitReason:      decision.Reason,
		HoldingDuration: holdingDuration,
		StrategyID:      open.StrategyID,
		WorkflowName:    open.WorkflowName,
		WorkflowVersion: open.WorkflowVersion,
		ParameterSetID:  open.ParameterSetID,
	})
	return nil
}

func (n *closedTradeStoreNode) Name() string {
	return "dagruntime.closed_trade_store." + n.id
}
func (n *closedTradeStoreNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{closedTradeOutputKey(n.tradeNodeID)}
}
func (n *closedTradeStoreNode) Provides() []artifact.AnyKey { return nil }
func (n *closedTradeStoreNode) Reads() []state.AnyKey {
	return []state.AnyKey{algotrade.StateClosedTrades}
}
func (n *closedTradeStoreNode) Writes() []state.AnyKey {
	return []state.AnyKey{algotrade.StateClosedTrades}
}
func (n *closedTradeStoreNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true, Idempotent: true}
}
func (n *closedTradeStoreNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = aw
	trade := artifact.MustGet(av, closedTradeOutputKey(n.tradeNodeID))
	if trade.TradeID == "" {
		return nil
	}
	closedTrades, _ := state.Get(txn, algotrade.StateClosedTrades)
	for _, item := range closedTrades.Items {
		if item.TradeID == trade.TradeID {
			return nil
		}
	}
	closedTrades.Items = append(closedTrades.Items, trade)
	state.StageWrite(txn, algotrade.StateClosedTrades, closedTrades)
	return nil
}

func (n *dailyPnLUpdateNode) Name() string { return "dagruntime.daily_pnl_update." + n.id }
func (n *dailyPnLUpdateNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{closedTradeOutputKey(n.tradeNodeID)}
}
func (n *dailyPnLUpdateNode) Provides() []artifact.AnyKey { return nil }
func (n *dailyPnLUpdateNode) Reads() []state.AnyKey {
	return []state.AnyKey{algotrade.StateDailyPnL}
}
func (n *dailyPnLUpdateNode) Writes() []state.AnyKey {
	return []state.AnyKey{algotrade.StateDailyPnL}
}
func (n *dailyPnLUpdateNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true, Idempotent: true}
}
func (n *dailyPnLUpdateNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = aw
	trade := artifact.MustGet(av, closedTradeOutputKey(n.tradeNodeID))
	if trade.TradeID == "" || trade.ExitTime.IsZero() {
		return nil
	}
	daily, _ := state.Get(txn, algotrade.StateDailyPnL)
	tradingDay := trade.ExitTime.Time().UTC().Format("2006-01-02")
	if daily.TradingDay != tradingDay {
		daily = algotrade.DailyPnLState{TradingDay: tradingDay}
	}
	daily.RealizedPnL += trade.NetPnL
	state.StageWrite(txn, algotrade.StateDailyPnL, daily)
	return nil
}

func (n *openPositionCloseNode) Name() string { return "dagruntime.open_position_close." + n.id }
func (n *openPositionCloseNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{closedTradeOutputKey(n.tradeNodeID)}
}
func (n *openPositionCloseNode) Provides() []artifact.AnyKey { return nil }
func (n *openPositionCloseNode) Reads() []state.AnyKey {
	return []state.AnyKey{algotrade.StateOpenPositions}
}
func (n *openPositionCloseNode) Writes() []state.AnyKey {
	return []state.AnyKey{algotrade.StateOpenPositions}
}
func (n *openPositionCloseNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true, Idempotent: true}
}
func (n *openPositionCloseNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = aw
	trade := artifact.MustGet(av, closedTradeOutputKey(n.tradeNodeID))
	if trade.PositionID == "" {
		return nil
	}
	openPositions, _ := state.Get(txn, algotrade.StateOpenPositions)
	filtered := openPositions.Items[:0]
	for _, item := range openPositions.Items {
		if item.PositionID != trade.PositionID {
			filtered = append(filtered, item)
		}
	}
	openPositions.Items = filtered
	state.StageWrite(txn, algotrade.StateOpenPositions, openPositions)
	return nil
}

func (n *strategySummaryUpdateNode) Name() string {
	return "dagruntime.strategy_summary_update." + n.id
}
func (n *strategySummaryUpdateNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{closedTradeOutputKey(n.tradeNodeID)}
}
func (n *strategySummaryUpdateNode) Provides() []artifact.AnyKey { return nil }
func (n *strategySummaryUpdateNode) Reads() []state.AnyKey {
	return []state.AnyKey{algotrade.StateClosedTrades, algotrade.StateStrategySummary}
}
func (n *strategySummaryUpdateNode) Writes() []state.AnyKey {
	return []state.AnyKey{algotrade.StateStrategySummary}
}
func (n *strategySummaryUpdateNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true, Idempotent: true}
}
func (n *strategySummaryUpdateNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = av
	_ = aw
	closedTrades, _ := state.Get(txn, algotrade.StateClosedTrades)
	if len(closedTrades.Items) == 0 {
		return nil
	}
	summary := summarizeClosedTrades(closedTrades.Items)
	state.StageWrite(txn, algotrade.StateStrategySummary, algotrade.StrategySummaryState{Summary: summary})
	return nil
}

func riskTrailingStopDistanceOutputKey(nodeID string) artifact.Key[marketdata.Price] {
	return artifact.Key[marketdata.Price]{
		Name:     fmt.Sprintf("%s.trailing_stop_distance", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.risk.trailing_stop_distance.%s.v1", nodeID),
	}
}

func positionSnapshotOutputKey(nodeID string) artifact.Key[algotrade.PositionSnapshot] {
	return artifact.Key[algotrade.PositionSnapshot]{
		Name:     fmt.Sprintf("%s.position_snapshot", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.position.snapshot.%s.v1", nodeID),
	}
}

func closedTradeOutputKey(nodeID string) artifact.Key[algotrade.ClosedTrade] {
	return artifact.Key[algotrade.ClosedTrade]{
		Name:     fmt.Sprintf("%s.closed_trade", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.closed_trade.%s.v1", nodeID),
	}
}

func closedTradeID(open algotrade.OpenPosition) string {
	if open.IntentID != "" {
		return fmt.Sprintf("trade:%s:1", open.IntentID)
	}
	return fmt.Sprintf("trade:%s:1", open.PositionID)
}

func closedTradeGrossPnL(side algotrade.PositionSide, entryPrice, exitPrice marketdata.Price, size float64) float64 {
	priceDelta := float64(exitPrice.Sub(entryPrice).Raw())
	switch side {
	case algotrade.PositionSideShort:
		priceDelta = -priceDelta
	}
	return priceDelta * size
}

func closedTradeNetPnL(grossPnL float64) float64 {
	return grossPnL
}

func summarizeClosedTrades(trades []algotrade.ClosedTrade) algotrade.StrategySummary {
	summary := algotrade.StrategySummary{}
	var winSum float64
	var lossSum float64
	for i, trade := range trades {
		if i == 0 {
			summary.StrategyID = trade.StrategyID
			summary.WorkflowName = trade.WorkflowName
			summary.WorkflowVersion = trade.WorkflowVersion
			summary.ParameterSetID = trade.ParameterSetID
		}
		summary.TradeCount++
		summary.TotalNetPnL += trade.NetPnL
		if trade.NetPnL > 0 {
			summary.WinCount++
			winSum += trade.NetPnL
		} else if trade.NetPnL < 0 {
			summary.LossCount++
			lossSum += -trade.NetPnL
		}
		if trade.ExitTime.After(summary.UpdatedAt) {
			summary.UpdatedAt = trade.ExitTime
		}
	}
	if summary.TradeCount > 0 {
		summary.WinRate = float64(summary.WinCount) / float64(summary.TradeCount)
	}
	if summary.WinCount > 0 {
		summary.AverageWin = winSum / float64(summary.WinCount)
	}
	if summary.LossCount > 0 {
		summary.AverageLoss = lossSum / float64(summary.LossCount)
	}
	if lossSum > 0 {
		summary.ProfitFactor = winSum / lossSum
	}
	return summary
}
