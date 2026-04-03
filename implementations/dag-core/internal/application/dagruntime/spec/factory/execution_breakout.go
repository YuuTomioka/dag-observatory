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

type ExecutionSubmitPaperOrderFactory struct{}

func (f *ExecutionSubmitPaperOrderFactory) Kind() string { return "execution_submit_paper_order" }

func (f *ExecutionSubmitPaperOrderFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"decision_node_id"}); err != nil {
		return nil, fmt.Errorf("execution_submit_paper_order node %q: %w", nodeSpec.ID, err)
	}
	decisionNodeID, err := requiredString(nodeSpec.Config, "decision_node_id")
	if err != nil {
		return nil, fmt.Errorf("execution_submit_paper_order node %q: %w", nodeSpec.ID, err)
	}
	return &executionSubmitPaperOrderNode{
		id:             nodeSpec.ID,
		decisionNodeID: decisionNodeID,
	}, nil
}

type ExecutionSubmitMarketOrderFactory struct{}

func (f *ExecutionSubmitMarketOrderFactory) Kind() string { return "execution_submit_market_order" }

func (f *ExecutionSubmitMarketOrderFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"decision_node_id"}); err != nil {
		return nil, fmt.Errorf("execution_submit_market_order node %q: %w", nodeSpec.ID, err)
	}
	decisionNodeID, err := requiredString(nodeSpec.Config, "decision_node_id")
	if err != nil {
		return nil, fmt.Errorf("execution_submit_market_order node %q: %w", nodeSpec.ID, err)
	}
	return &executionSubmitMarketOrderNode{
		id:             nodeSpec.ID,
		decisionNodeID: decisionNodeID,
	}, nil
}

type ExecutionConfirmFillFactory struct{}

func (f *ExecutionConfirmFillFactory) Kind() string { return "execution_confirm_fill" }

func (f *ExecutionConfirmFillFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"order_request_node_id"}); err != nil {
		return nil, fmt.Errorf("execution_confirm_fill node %q: %w", nodeSpec.ID, err)
	}
	orderRequestNodeID, err := requiredString(nodeSpec.Config, "order_request_node_id")
	if err != nil {
		return nil, fmt.Errorf("execution_confirm_fill node %q: %w", nodeSpec.ID, err)
	}
	return &executionConfirmFillNode{
		id:                 nodeSpec.ID,
		orderRequestNodeID: orderRequestNodeID,
	}, nil
}

type executionSubmitPaperOrderNode struct {
	id             string
	decisionNodeID string
}

func (n *executionSubmitPaperOrderNode) Name() string {
	return "dagruntime.execution_submit_paper_order." + n.id
}
func (n *executionSubmitPaperOrderNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{signalDecisionOutputKey(n.decisionNodeID)}
}
func (n *executionSubmitPaperOrderNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{paperExecutionResultOutputKey(n.id)}
}
func (n *executionSubmitPaperOrderNode) Reads() []state.AnyKey  { return nil }
func (n *executionSubmitPaperOrderNode) Writes() []state.AnyKey { return nil }
func (n *executionSubmitPaperOrderNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}

func (n *executionSubmitPaperOrderNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	decision := artifact.MustGet(av, signalDecisionOutputKey(n.decisionNodeID))
	result := algotrade.PaperExecutionResult{
		Submitted: false,
		Action:    decision.Action,
		Size:      0,
		Reason:    decision.Reason,
	}
	if decision.Action == algotrade.OrderActionBuy && decision.PositionSize > 0 {
		result.Submitted = true
		result.Size = decision.PositionSize
	}
	artifact.Set(aw, paperExecutionResultOutputKey(n.id), result)
	return nil
}

type executionSubmitMarketOrderNode struct {
	id             string
	decisionNodeID string
}

func (n *executionSubmitMarketOrderNode) Name() string {
	return "dagruntime.execution_submit_market_order." + n.id
}
func (n *executionSubmitMarketOrderNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{signalDecisionOutputKey(n.decisionNodeID)}
}
func (n *executionSubmitMarketOrderNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{executionMarketOrderRequestOutputKey(n.id)}
}
func (n *executionSubmitMarketOrderNode) Reads() []state.AnyKey {
	return []state.AnyKey{algotrade.StatePendingOrders}
}
func (n *executionSubmitMarketOrderNode) Writes() []state.AnyKey {
	return []state.AnyKey{algotrade.StatePendingOrders}
}
func (n *executionSubmitMarketOrderNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{
		Deterministic: false,
		Idempotent:    true,
		SideEffect:    true,
	}
}

func (n *executionSubmitMarketOrderNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	decision := artifact.MustGet(av, signalDecisionOutputKey(n.decisionNodeID))
	req := algotrade.MarketOrderRequest{
		Submitted: false,
		Action:    decision.Action,
		Size:      0,
		Reason:    decision.Reason,
		SourceID: fmt.Sprintf(
			"%s:%s:%0.8f:%d",
			decision.Action,
			decision.Reason,
			decision.PositionSize,
			decision.StopLossDistance.Raw(),
		),
	}
	if decision.Action != algotrade.OrderActionBuy || decision.PositionSize <= 0 {
		artifact.Set(aw, executionMarketOrderRequestOutputKey(n.id), req)
		return nil
	}

	pending, _ := state.Get(txn, algotrade.StatePendingOrders)
	for _, item := range pending.Items {
		if item.SourceID == req.SourceID {
			req.Submitted = true
			req.OrderID = item.OrderID
			req.Size = item.Request.PositionSize
			req.Reason = "pending_order_exists"
			artifact.Set(aw, executionMarketOrderRequestOutputKey(n.id), req)
			return nil
		}
	}

	symbol, _ := artifact.Get(av, usecase.InputKeySymbol)
	req.Submitted = true
	req.OrderID = fmt.Sprintf("%s-%d", n.id, len(pending.Items)+1)
	req.Size = decision.PositionSize
	req.Reason = "submitted"
	pending.Items = append(pending.Items, algotrade.PendingOrder{
		OrderID:  req.OrderID,
		Symbol:   symbol,
		Request:  decision,
		SourceID: req.SourceID,
	})
	state.StageWrite(txn, algotrade.StatePendingOrders, pending)
	artifact.Set(aw, executionMarketOrderRequestOutputKey(n.id), req)
	return nil
}

type executionConfirmFillNode struct {
	id                 string
	orderRequestNodeID string
}

func (n *executionConfirmFillNode) Name() string {
	return "dagruntime.execution_confirm_fill." + n.id
}
func (n *executionConfirmFillNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{executionMarketOrderRequestOutputKey(n.orderRequestNodeID)}
}
func (n *executionConfirmFillNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{executionFillResultOutputKey(n.id)}
}
func (n *executionConfirmFillNode) Reads() []state.AnyKey {
	return []state.AnyKey{
		algotrade.StatePendingOrders,
		algotrade.StateOpenPositions,
	}
}
func (n *executionConfirmFillNode) Writes() []state.AnyKey {
	return []state.AnyKey{
		algotrade.StatePendingOrders,
		algotrade.StateOpenPositions,
	}
}
func (n *executionConfirmFillNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{
		Deterministic: false,
		Idempotent:    true,
		SideEffect:    true,
	}
}

func (n *executionConfirmFillNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	request := artifact.MustGet(av, executionMarketOrderRequestOutputKey(n.orderRequestNodeID))
	result := algotrade.FillResult{
		Filled:  false,
		OrderID: request.OrderID,
		Action:  request.Action,
		Size:    request.Size,
		Reason:  request.Reason,
	}
	if !request.Submitted || request.OrderID == "" {
		artifact.Set(aw, executionFillResultOutputKey(n.id), result)
		return nil
	}

	pending, _ := state.Get(txn, algotrade.StatePendingOrders)
	pendingIndex := -1
	var matched algotrade.PendingOrder
	for i, item := range pending.Items {
		if item.OrderID == request.OrderID {
			pendingIndex = i
			matched = item
			break
		}
	}
	if pendingIndex < 0 {
		result.Reason = "pending_order_not_found"
		artifact.Set(aw, executionFillResultOutputKey(n.id), result)
		return nil
	}

	pending.Items = append(pending.Items[:pendingIndex], pending.Items[pendingIndex+1:]...)
	state.StageWrite(txn, algotrade.StatePendingOrders, pending)

	openPositions, _ := state.Get(txn, algotrade.StateOpenPositions)
	if matched.Request.Action == algotrade.OrderActionBuy && matched.Request.PositionSize > 0 {
		snapshot := algotrade.PositionSnapshot{
			HasPosition: true,
			Side:        algotrade.PositionSideLong,
			Size:        matched.Request.PositionSize,
			EntryPrice:  marketdata.Price(0),
		}
		if bars, ok := artifact.Get(av, usecase.InputKeyMarketOHLCVBars); ok && len(bars) > 0 {
			snapshot.EntryPrice = bars[len(bars)-1].Close
		}
		replaced := false
		for i := range openPositions.Items {
			if openPositions.Items[i].Symbol == matched.Symbol {
				openPositions.Items[i] = algotrade.OpenPosition{
					PositionID: matched.OrderID,
					Symbol:     matched.Symbol,
					Snapshot:   snapshot,
				}
				replaced = true
				break
			}
		}
		if !replaced {
			openPositions.Items = append(openPositions.Items, algotrade.OpenPosition{
				PositionID: matched.OrderID,
				Symbol:     matched.Symbol,
				Snapshot:   snapshot,
			})
		}
		state.StageWrite(txn, algotrade.StateOpenPositions, openPositions)
		result.Filled = true
		result.Reason = "filled"
		artifact.Set(aw, executionFillResultOutputKey(n.id), result)
		return nil
	}

	result.Reason = "unsupported_action"
	artifact.Set(aw, executionFillResultOutputKey(n.id), result)
	return nil
}

func paperExecutionResultOutputKey(nodeID string) artifact.Key[algotrade.PaperExecutionResult] {
	return artifact.Key[algotrade.PaperExecutionResult]{
		Name:     fmt.Sprintf("%s.paper_execution_result", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.execution.paper.%s.v1", nodeID),
	}
}

func executionMarketOrderRequestOutputKey(nodeID string) artifact.Key[algotrade.MarketOrderRequest] {
	return artifact.Key[algotrade.MarketOrderRequest]{
		Name:     fmt.Sprintf("%s.market_order_request", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.execution.market_order_request.%s.v1", nodeID),
	}
}

func executionFillResultOutputKey(nodeID string) artifact.Key[algotrade.FillResult] {
	return artifact.Key[algotrade.FillResult]{
		Name:     fmt.Sprintf("%s.fill_result", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.execution.fill_result.%s.v1", nodeID),
	}
}
