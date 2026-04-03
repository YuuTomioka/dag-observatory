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
	artifact.Set(aw, positionSnapshotOutputKey(n.id), openPositions.Items[0].Snapshot)
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
