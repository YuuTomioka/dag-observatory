package factory

import (
	"context"
	"fmt"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	"dag-observatory/dag-core/internal/domain/algotrade"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type ObservabilityEmitSignalDecisionFactory struct{}

func (f *ObservabilityEmitSignalDecisionFactory) Kind() string {
	return "observability_emit_signal_decision"
}

func (f *ObservabilityEmitSignalDecisionFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"decision_node_id"}); err != nil {
		return nil, fmt.Errorf("observability_emit_signal_decision node %q: %w", nodeSpec.ID, err)
	}
	decisionNodeID, err := requiredString(nodeSpec.Config, "decision_node_id")
	if err != nil {
		return nil, fmt.Errorf("observability_emit_signal_decision node %q: %w", nodeSpec.ID, err)
	}
	return &observabilityEmitSignalDecisionNode{
		id:             nodeSpec.ID,
		decisionNodeID: decisionNodeID,
	}, nil
}

type ObservabilityEmitOrderDecisionFactory struct{}

func (f *ObservabilityEmitOrderDecisionFactory) Kind() string {
	return "observability_emit_order_decision"
}

func (f *ObservabilityEmitOrderDecisionFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"decision_node_id"}); err != nil {
		return nil, fmt.Errorf("observability_emit_order_decision node %q: %w", nodeSpec.ID, err)
	}
	decisionNodeID, err := requiredString(nodeSpec.Config, "decision_node_id")
	if err != nil {
		return nil, fmt.Errorf("observability_emit_order_decision node %q: %w", nodeSpec.ID, err)
	}
	return &observabilityEmitOrderDecisionNode{
		id:             nodeSpec.ID,
		decisionNodeID: decisionNodeID,
	}, nil
}

type ObservabilityEmitPositionEventFactory struct{}

func (f *ObservabilityEmitPositionEventFactory) Kind() string {
	return "observability_emit_position_event"
}

func (f *ObservabilityEmitPositionEventFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"position_node_id"}); err != nil {
		return nil, fmt.Errorf("observability_emit_position_event node %q: %w", nodeSpec.ID, err)
	}
	positionNodeID, err := requiredString(nodeSpec.Config, "position_node_id")
	if err != nil {
		return nil, fmt.Errorf("observability_emit_position_event node %q: %w", nodeSpec.ID, err)
	}
	return &observabilityEmitPositionEventNode{
		id:             nodeSpec.ID,
		positionNodeID: positionNodeID,
	}, nil
}

type observabilityEmitSignalDecisionNode struct {
	id             string
	decisionNodeID string
}

func (n *observabilityEmitSignalDecisionNode) Name() string {
	return "dagruntime.observability_emit_signal_decision." + n.id
}
func (n *observabilityEmitSignalDecisionNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{signalDecisionOutputKey(n.decisionNodeID)}
}
func (n *observabilityEmitSignalDecisionNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{observabilitySignalDecisionOutputKey(n.id)}
}
func (n *observabilityEmitSignalDecisionNode) Reads() []state.AnyKey  { return nil }
func (n *observabilityEmitSignalDecisionNode) Writes() []state.AnyKey { return nil }
func (n *observabilityEmitSignalDecisionNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}

func (n *observabilityEmitSignalDecisionNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	decision := artifact.MustGet(av, signalDecisionOutputKey(n.decisionNodeID))
	artifact.Set(aw, observabilitySignalDecisionOutputKey(n.id), observabilitySignalDecision{
		Action:       decision.Action,
		Reason:       decision.Reason,
		Allowed:      decision.Action == algotrade.OrderActionBuy,
		PositionSize: decision.PositionSize,
	})
	return nil
}

type observabilityEmitOrderDecisionNode struct {
	id             string
	decisionNodeID string
}

func (n *observabilityEmitOrderDecisionNode) Name() string {
	return "dagruntime.observability_emit_order_decision." + n.id
}
func (n *observabilityEmitOrderDecisionNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{signalDecisionOutputKey(n.decisionNodeID)}
}
func (n *observabilityEmitOrderDecisionNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{observabilityOrderDecisionOutputKey(n.id)}
}
func (n *observabilityEmitOrderDecisionNode) Reads() []state.AnyKey  { return nil }
func (n *observabilityEmitOrderDecisionNode) Writes() []state.AnyKey { return nil }
func (n *observabilityEmitOrderDecisionNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *observabilityEmitOrderDecisionNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	decision := artifact.MustGet(av, signalDecisionOutputKey(n.decisionNodeID))
	artifact.Set(aw, observabilityOrderDecisionOutputKey(n.id), observabilityOrderDecision{
		Action:       decision.Action,
		Reason:       decision.Reason,
		PositionSize: decision.PositionSize,
	})
	return nil
}

type observabilityEmitPositionEventNode struct {
	id             string
	positionNodeID string
}

func (n *observabilityEmitPositionEventNode) Name() string {
	return "dagruntime.observability_emit_position_event." + n.id
}
func (n *observabilityEmitPositionEventNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{positionSnapshotOutputKey(n.positionNodeID)}
}
func (n *observabilityEmitPositionEventNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{observabilityPositionEventOutputKey(n.id)}
}
func (n *observabilityEmitPositionEventNode) Reads() []state.AnyKey  { return nil }
func (n *observabilityEmitPositionEventNode) Writes() []state.AnyKey { return nil }
func (n *observabilityEmitPositionEventNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *observabilityEmitPositionEventNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	position := artifact.MustGet(av, positionSnapshotOutputKey(n.positionNodeID))
	artifact.Set(aw, observabilityPositionEventOutputKey(n.id), observabilityPositionEvent{
		HasPosition: position.HasPosition,
		Side:        position.Side,
		Size:        position.Size,
	})
	return nil
}

func observabilitySignalDecisionOutputKey(nodeID string) artifact.Key[observabilitySignalDecision] {
	return artifact.Key[observabilitySignalDecision]{
		Name:     fmt.Sprintf("%s.observability_signal_decision", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.observability.signal_decision.%s.v1", nodeID),
	}
}

func observabilityOrderDecisionOutputKey(nodeID string) artifact.Key[observabilityOrderDecision] {
	return artifact.Key[observabilityOrderDecision]{
		Name:     fmt.Sprintf("%s.observability_order_decision", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.observability.order_decision.%s.v1", nodeID),
	}
}

func observabilityPositionEventOutputKey(nodeID string) artifact.Key[observabilityPositionEvent] {
	return artifact.Key[observabilityPositionEvent]{
		Name:     fmt.Sprintf("%s.observability_position_event", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.observability.position_event.%s.v1", nodeID),
	}
}
