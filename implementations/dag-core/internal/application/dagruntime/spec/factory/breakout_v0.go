package factory

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/algotrade"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

type FeatureATRFactory struct{}

func (f *FeatureATRFactory) Kind() string { return "feature_atr" }

func (f *FeatureATRFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"window"}); err != nil {
		return nil, fmt.Errorf("feature_atr node %q: %w", nodeSpec.ID, err)
	}
	window, err := requiredInt(nodeSpec.Config, "window")
	if err != nil {
		return nil, fmt.Errorf("feature_atr node %q: %w", nodeSpec.ID, err)
	}
	if window <= 0 {
		return nil, fmt.Errorf("feature_atr node %q: config.window must be > 0", nodeSpec.ID)
	}
	return &featureATRNode{
		id:     nodeSpec.ID,
		window: window,
	}, nil
}

type FeatureRangeHighFactory struct{}

func (f *FeatureRangeHighFactory) Kind() string { return "feature_range_high" }

func (f *FeatureRangeHighFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"window"}); err != nil {
		return nil, fmt.Errorf("feature_range_high node %q: %w", nodeSpec.ID, err)
	}
	window, err := requiredInt(nodeSpec.Config, "window")
	if err != nil {
		return nil, fmt.Errorf("feature_range_high node %q: %w", nodeSpec.ID, err)
	}
	if window <= 0 {
		return nil, fmt.Errorf("feature_range_high node %q: config.window must be > 0", nodeSpec.ID)
	}
	return &featureRangeHighNode{
		id:     nodeSpec.ID,
		window: window,
	}, nil
}

type FeatureRangeLowFactory struct{}

func (f *FeatureRangeLowFactory) Kind() string { return "feature_range_low" }

func (f *FeatureRangeLowFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"window"}); err != nil {
		return nil, fmt.Errorf("feature_range_low node %q: %w", nodeSpec.ID, err)
	}
	window, err := requiredInt(nodeSpec.Config, "window")
	if err != nil {
		return nil, fmt.Errorf("feature_range_low node %q: %w", nodeSpec.ID, err)
	}
	if window <= 0 {
		return nil, fmt.Errorf("feature_range_low node %q: config.window must be > 0", nodeSpec.ID)
	}
	return &featureRangeLowNode{
		id:     nodeSpec.ID,
		window: window,
	}, nil
}

type FeatureSpreadFactory struct{}

func (f *FeatureSpreadFactory) Kind() string { return "feature_spread" }

func (f *FeatureSpreadFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, nil); err != nil {
		return nil, fmt.Errorf("feature_spread node %q: %w", nodeSpec.ID, err)
	}
	return &featureSpreadNode{id: nodeSpec.ID}, nil
}

type FeatureSessionStateFactory struct{}

func (f *FeatureSessionStateFactory) Kind() string { return "feature_session_state" }

func (f *FeatureSessionStateFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, nil); err != nil {
		return nil, fmt.Errorf("feature_session_state node %q: %w", nodeSpec.ID, err)
	}
	return &featureSessionStateNode{id: nodeSpec.ID}, nil
}

type FeatureHigherTFTrendFactory struct{}

func (f *FeatureHigherTFTrendFactory) Kind() string { return "feature_higher_tf_trend" }

func (f *FeatureHigherTFTrendFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, nil); err != nil {
		return nil, fmt.Errorf("feature_higher_tf_trend node %q: %w", nodeSpec.ID, err)
	}
	return &featureHigherTFTrendNode{id: nodeSpec.ID}, nil
}

type SignalBreakoutLongFactory struct{}

func (f *SignalBreakoutLongFactory) Kind() string { return "signal_breakout_long" }

func (f *SignalBreakoutLongFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"range_node_id"}); err != nil {
		return nil, fmt.Errorf("signal_breakout_long node %q: %w", nodeSpec.ID, err)
	}
	rangeNodeID, err := requiredString(nodeSpec.Config, "range_node_id")
	if err != nil {
		return nil, fmt.Errorf("signal_breakout_long node %q: %w", nodeSpec.ID, err)
	}
	return &signalBreakoutLongNode{
		id:          nodeSpec.ID,
		rangeNodeID: rangeNodeID,
	}, nil
}

type SignalExitBasicFactory struct{}

func (f *SignalExitBasicFactory) Kind() string { return "signal_exit_basic" }

func (f *SignalExitBasicFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"position_node_id", "stop_loss_node_id"}); err != nil {
		return nil, fmt.Errorf("signal_exit_basic node %q: %w", nodeSpec.ID, err)
	}
	positionNodeID, err := requiredString(nodeSpec.Config, "position_node_id")
	if err != nil {
		return nil, fmt.Errorf("signal_exit_basic node %q: %w", nodeSpec.ID, err)
	}
	stopLossNodeID, err := requiredString(nodeSpec.Config, "stop_loss_node_id")
	if err != nil {
		return nil, fmt.Errorf("signal_exit_basic node %q: %w", nodeSpec.ID, err)
	}
	return &signalExitBasicNode{
		id:             nodeSpec.ID,
		positionNodeID: positionNodeID,
		stopLossNodeID: stopLossNodeID,
	}, nil
}

type FilterSessionFactory struct{}

func (f *FilterSessionFactory) Kind() string { return "filter_session" }

func (f *FilterSessionFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"signal_node_id", "allowed_sessions"}); err != nil {
		return nil, fmt.Errorf("filter_session node %q: %w", nodeSpec.ID, err)
	}
	signalNodeID, err := requiredString(nodeSpec.Config, "signal_node_id")
	if err != nil {
		return nil, fmt.Errorf("filter_session node %q: %w", nodeSpec.ID, err)
	}
	sessions := optionalStringSlice(nodeSpec.Config, "allowed_sessions", []string{"tokyo", "london", "newyork"})
	if len(sessions) == 0 {
		return nil, fmt.Errorf("filter_session node %q: config.allowed_sessions must not be empty", nodeSpec.ID)
	}
	return &filterSessionNode{
		id:              nodeSpec.ID,
		signalNodeID:    signalNodeID,
		allowedSessions: sessions,
	}, nil
}

type FilterSpreadFactory struct{}

func (f *FilterSpreadFactory) Kind() string { return "filter_spread" }

func (f *FilterSpreadFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"allowed_node_id", "max_spread_bps"}); err != nil {
		return nil, fmt.Errorf("filter_spread node %q: %w", nodeSpec.ID, err)
	}
	allowedNodeID, err := requiredString(nodeSpec.Config, "allowed_node_id")
	if err != nil {
		return nil, fmt.Errorf("filter_spread node %q: %w", nodeSpec.ID, err)
	}
	maxSpreadBps, err := requiredFloat(nodeSpec.Config, "max_spread_bps")
	if err != nil {
		return nil, fmt.Errorf("filter_spread node %q: %w", nodeSpec.ID, err)
	}
	if maxSpreadBps <= 0 {
		return nil, fmt.Errorf("filter_spread node %q: config.max_spread_bps must be > 0", nodeSpec.ID)
	}
	return &filterSpreadNode{
		id:            nodeSpec.ID,
		allowedNodeID: allowedNodeID,
		maxSpreadBps:  maxSpreadBps,
	}, nil
}

type FilterDailyLossLimitFactory struct{}

func (f *FilterDailyLossLimitFactory) Kind() string { return "filter_daily_loss_limit" }

func (f *FilterDailyLossLimitFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"allowed_node_id", "max_daily_loss"}); err != nil {
		return nil, fmt.Errorf("filter_daily_loss_limit node %q: %w", nodeSpec.ID, err)
	}
	allowedNodeID, err := requiredString(nodeSpec.Config, "allowed_node_id")
	if err != nil {
		return nil, fmt.Errorf("filter_daily_loss_limit node %q: %w", nodeSpec.ID, err)
	}
	maxDailyLoss := optionalFloat(nodeSpec.Config, "max_daily_loss", 0)
	if maxDailyLoss < 0 {
		return nil, fmt.Errorf("filter_daily_loss_limit node %q: config.max_daily_loss must be >= 0", nodeSpec.ID)
	}
	return &filterDailyLossLimitNode{
		id:            nodeSpec.ID,
		allowedNodeID: allowedNodeID,
		maxDailyLoss:  maxDailyLoss,
	}, nil
}

type FilterEconomicEventFactory struct{}

func (f *FilterEconomicEventFactory) Kind() string { return "filter_economic_event" }

func (f *FilterEconomicEventFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"allowed_node_id", "block_severity"}); err != nil {
		return nil, fmt.Errorf("filter_economic_event node %q: %w", nodeSpec.ID, err)
	}
	allowedNodeID, err := requiredString(nodeSpec.Config, "allowed_node_id")
	if err != nil {
		return nil, fmt.Errorf("filter_economic_event node %q: %w", nodeSpec.ID, err)
	}
	blockSeverity := optionalString(nodeSpec.Config, "block_severity", "high")
	return &filterEconomicEventNode{
		id:            nodeSpec.ID,
		allowedNodeID: allowedNodeID,
		blockSeverity: strings.ToLower(strings.TrimSpace(blockSeverity)),
	}, nil
}

type FilterHigherTFAlignmentFactory struct{}

func (f *FilterHigherTFAlignmentFactory) Kind() string { return "filter_higher_tf_alignment" }

func (f *FilterHigherTFAlignmentFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"allowed_node_id", "trend_node_id", "required_direction"}); err != nil {
		return nil, fmt.Errorf("filter_higher_tf_alignment node %q: %w", nodeSpec.ID, err)
	}
	allowedNodeID, err := requiredString(nodeSpec.Config, "allowed_node_id")
	if err != nil {
		return nil, fmt.Errorf("filter_higher_tf_alignment node %q: %w", nodeSpec.ID, err)
	}
	trendNodeID, err := requiredString(nodeSpec.Config, "trend_node_id")
	if err != nil {
		return nil, fmt.Errorf("filter_higher_tf_alignment node %q: %w", nodeSpec.ID, err)
	}
	requiredDirection := optionalString(nodeSpec.Config, "required_direction", "up")
	requiredDirection = strings.ToLower(strings.TrimSpace(requiredDirection))
	if requiredDirection == "" {
		requiredDirection = "up"
	}
	return &filterHigherTFAlignmentNode{
		id:                nodeSpec.ID,
		allowedNodeID:     allowedNodeID,
		trendNodeID:       trendNodeID,
		requiredDirection: requiredDirection,
	}, nil
}

type RiskPositionSizingFactory struct{}

func (f *RiskPositionSizingFactory) Kind() string { return "risk_position_sizing" }

func (f *RiskPositionSizingFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"allowed_node_id", "atr_node_id", "risk_rate", "stop_atr_multiple"}); err != nil {
		return nil, fmt.Errorf("risk_position_sizing node %q: %w", nodeSpec.ID, err)
	}
	allowedNodeID, err := requiredString(nodeSpec.Config, "allowed_node_id")
	if err != nil {
		return nil, fmt.Errorf("risk_position_sizing node %q: %w", nodeSpec.ID, err)
	}
	atrNodeID, err := requiredString(nodeSpec.Config, "atr_node_id")
	if err != nil {
		return nil, fmt.Errorf("risk_position_sizing node %q: %w", nodeSpec.ID, err)
	}
	riskRate, err := requiredFloat(nodeSpec.Config, "risk_rate")
	if err != nil {
		return nil, fmt.Errorf("risk_position_sizing node %q: %w", nodeSpec.ID, err)
	}
	if riskRate <= 0 || riskRate > 1 {
		return nil, fmt.Errorf("risk_position_sizing node %q: config.risk_rate must be in (0,1]", nodeSpec.ID)
	}
	stopATRMultiple := optionalFloat(nodeSpec.Config, "stop_atr_multiple", 1.5)
	if stopATRMultiple <= 0 {
		return nil, fmt.Errorf("risk_position_sizing node %q: config.stop_atr_multiple must be > 0", nodeSpec.ID)
	}
	return &riskPositionSizingNode{
		id:              nodeSpec.ID,
		allowedNodeID:   allowedNodeID,
		atrNodeID:       atrNodeID,
		riskRate:        riskRate,
		stopATRMultiple: stopATRMultiple,
	}, nil
}

type RiskMaxPositionsCheckFactory struct{}

func (f *RiskMaxPositionsCheckFactory) Kind() string { return "risk_max_positions_check" }

func (f *RiskMaxPositionsCheckFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"allowed_node_id", "max_positions"}); err != nil {
		return nil, fmt.Errorf("risk_max_positions_check node %q: %w", nodeSpec.ID, err)
	}
	allowedNodeID, err := requiredString(nodeSpec.Config, "allowed_node_id")
	if err != nil {
		return nil, fmt.Errorf("risk_max_positions_check node %q: %w", nodeSpec.ID, err)
	}
	maxPositions := optionalInt(nodeSpec.Config, "max_positions", 1)
	if maxPositions <= 0 {
		return nil, fmt.Errorf("risk_max_positions_check node %q: config.max_positions must be > 0", nodeSpec.ID)
	}
	return &riskMaxPositionsCheckNode{
		id:            nodeSpec.ID,
		allowedNodeID: allowedNodeID,
		maxPositions:  maxPositions,
	}, nil
}

type RiskStopLossFromATRFactory struct{}

func (f *RiskStopLossFromATRFactory) Kind() string { return "risk_stop_loss_from_atr" }

func (f *RiskStopLossFromATRFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"atr_node_id", "atr_multiple"}); err != nil {
		return nil, fmt.Errorf("risk_stop_loss_from_atr node %q: %w", nodeSpec.ID, err)
	}
	atrNodeID, err := requiredString(nodeSpec.Config, "atr_node_id")
	if err != nil {
		return nil, fmt.Errorf("risk_stop_loss_from_atr node %q: %w", nodeSpec.ID, err)
	}
	atrMultiple := optionalFloat(nodeSpec.Config, "atr_multiple", 1.5)
	if atrMultiple <= 0 {
		return nil, fmt.Errorf("risk_stop_loss_from_atr node %q: config.atr_multiple must be > 0", nodeSpec.ID)
	}
	return &riskStopLossFromATRNode{
		id:          nodeSpec.ID,
		atrNodeID:   atrNodeID,
		atrMultiple: atrMultiple,
	}, nil
}

type RiskTakeProfitFromRRFactory struct{}

func (f *RiskTakeProfitFromRRFactory) Kind() string { return "risk_take_profit_from_rr" }

func (f *RiskTakeProfitFromRRFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"stop_loss_node_id", "rr_ratio"}); err != nil {
		return nil, fmt.Errorf("risk_take_profit_from_rr node %q: %w", nodeSpec.ID, err)
	}
	stopLossNodeID, err := requiredString(nodeSpec.Config, "stop_loss_node_id")
	if err != nil {
		return nil, fmt.Errorf("risk_take_profit_from_rr node %q: %w", nodeSpec.ID, err)
	}
	rrRatio, err := requiredFloat(nodeSpec.Config, "rr_ratio")
	if err != nil {
		return nil, fmt.Errorf("risk_take_profit_from_rr node %q: %w", nodeSpec.ID, err)
	}
	if rrRatio <= 0 {
		return nil, fmt.Errorf("risk_take_profit_from_rr node %q: config.rr_ratio must be > 0", nodeSpec.ID)
	}
	return &riskTakeProfitFromRRNode{
		id:             nodeSpec.ID,
		stopLossNodeID: stopLossNodeID,
		rrRatio:        rrRatio,
	}, nil
}

type SignalDecisionMapperFactory struct{}

func (f *SignalDecisionMapperFactory) Kind() string { return "signal_decision_mapper" }

func (f *SignalDecisionMapperFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"allowed_node_id", "sizing_node_id"}); err != nil {
		return nil, fmt.Errorf("signal_decision_mapper node %q: %w", nodeSpec.ID, err)
	}
	allowedNodeID, err := requiredString(nodeSpec.Config, "allowed_node_id")
	if err != nil {
		return nil, fmt.Errorf("signal_decision_mapper node %q: %w", nodeSpec.ID, err)
	}
	sizingNodeID, err := requiredString(nodeSpec.Config, "sizing_node_id")
	if err != nil {
		return nil, fmt.Errorf("signal_decision_mapper node %q: %w", nodeSpec.ID, err)
	}
	return &signalDecisionMapperNode{
		id:            nodeSpec.ID,
		allowedNodeID: allowedNodeID,
		sizingNodeID:  sizingNodeID,
	}, nil
}

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

type featureATRNode struct {
	id     string
	window int
}

func (n *featureATRNode) Name() string { return "dagruntime.feature_atr." + n.id }
func (n *featureATRNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{usecase.InputKeyMarketOHLCVBars}
}
func (n *featureATRNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{featureATROutputKey(n.id)}
}
func (n *featureATRNode) Reads() []state.AnyKey    { return nil }
func (n *featureATRNode) Writes() []state.AnyKey   { return nil }
func (n *featureATRNode) Spec() node.ExecutionSpec { return node.ExecutionSpec{Deterministic: true} }

func (n *featureATRNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	bars := artifact.MustGet(av, usecase.InputKeyMarketOHLCVBars)
	if len(bars) == 0 {
		return fmt.Errorf("feature_atr node %q: market.ohlcv_bars is empty", n.id)
	}
	window := n.window
	if window > len(bars) {
		window = len(bars)
	}
	start := len(bars) - window
	sumTR := int64(0)
	for i := start; i < len(bars); i++ {
		bar := bars[i]
		tr := bar.High.Sub(bar.Low).Abs().Raw()
		if i > 0 {
			prevClose := bars[i-1].Close
			d1 := bar.High.Sub(prevClose).Abs().Raw()
			d2 := bar.Low.Sub(prevClose).Abs().Raw()
			if d1 > tr {
				tr = d1
			}
			if d2 > tr {
				tr = d2
			}
		}
		sumTR += tr
	}
	atr := marketdata.NewPriceFromRaw(sumTR / int64(window))
	artifact.Set(aw, featureATROutputKey(n.id), algotrade.FeatureATR{Value: atr})
	return nil
}

type featureRangeHighNode struct {
	id     string
	window int
}

func (n *featureRangeHighNode) Name() string { return "dagruntime.feature_range_high." + n.id }
func (n *featureRangeHighNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{usecase.InputKeyMarketOHLCVBars}
}
func (n *featureRangeHighNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{featureRangeHighOutputKey(n.id)}
}
func (n *featureRangeHighNode) Reads() []state.AnyKey  { return nil }
func (n *featureRangeHighNode) Writes() []state.AnyKey { return nil }
func (n *featureRangeHighNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}

func (n *featureRangeHighNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	bars := artifact.MustGet(av, usecase.InputKeyMarketOHLCVBars)
	if len(bars) < 2 {
		return fmt.Errorf("feature_range_high node %q: market.ohlcv_bars requires at least 2 bars", n.id)
	}
	previousBars := bars[:len(bars)-1]
	window := n.window
	if window > len(previousBars) {
		window = len(previousBars)
	}
	start := len(previousBars) - window
	maxHigh := previousBars[start].High
	for _, bar := range previousBars[start+1:] {
		if bar.High.Gt(maxHigh) {
			maxHigh = bar.High
		}
	}
	artifact.Set(aw, featureRangeHighOutputKey(n.id), algotrade.FeatureRangeHigh{Value: maxHigh})
	return nil
}

type featureRangeLowNode struct {
	id     string
	window int
}

func (n *featureRangeLowNode) Name() string { return "dagruntime.feature_range_low." + n.id }
func (n *featureRangeLowNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{usecase.InputKeyMarketOHLCVBars}
}
func (n *featureRangeLowNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{featureRangeLowOutputKey(n.id)}
}
func (n *featureRangeLowNode) Reads() []state.AnyKey  { return nil }
func (n *featureRangeLowNode) Writes() []state.AnyKey { return nil }
func (n *featureRangeLowNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}

func (n *featureRangeLowNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	bars := artifact.MustGet(av, usecase.InputKeyMarketOHLCVBars)
	if len(bars) < 2 {
		return fmt.Errorf("feature_range_low node %q: market.ohlcv_bars requires at least 2 bars", n.id)
	}
	previousBars := bars[:len(bars)-1]
	window := n.window
	if window > len(previousBars) {
		window = len(previousBars)
	}
	start := len(previousBars) - window
	minLow := previousBars[start].Low
	for _, bar := range previousBars[start+1:] {
		if bar.Low.Lt(minLow) {
			minLow = bar.Low
		}
	}
	artifact.Set(aw, featureRangeLowOutputKey(n.id), algotrade.FeatureRangeLow{Value: minLow})
	return nil
}

type featureSpreadNode struct {
	id string
}

func (n *featureSpreadNode) Name() string { return "dagruntime.feature_spread." + n.id }
func (n *featureSpreadNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{usecase.InputKeyMarketSpreadBps}
}
func (n *featureSpreadNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{featureSpreadOutputKey(n.id)}
}
func (n *featureSpreadNode) Reads() []state.AnyKey  { return nil }
func (n *featureSpreadNode) Writes() []state.AnyKey { return nil }
func (n *featureSpreadNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *featureSpreadNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	spread := artifact.MustGet(av, usecase.InputKeyMarketSpreadBps)
	if spread < 0 {
		return fmt.Errorf("feature_spread node %q: market.spread_bps must be >= 0", n.id)
	}
	artifact.Set(aw, featureSpreadOutputKey(n.id), algotrade.FeatureSpread{Bps: spread})
	return nil
}

type featureSessionStateNode struct {
	id string
}

type featureHigherTFTrendNode struct {
	id string
}

func (n *featureHigherTFTrendNode) Name() string { return "dagruntime.feature_higher_tf_trend." + n.id }
func (n *featureHigherTFTrendNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{inputKeyMarketOHLCVBarsH1}
}
func (n *featureHigherTFTrendNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{featureHigherTFTrendOutputKey(n.id)}
}
func (n *featureHigherTFTrendNode) Reads() []state.AnyKey  { return nil }
func (n *featureHigherTFTrendNode) Writes() []state.AnyKey { return nil }
func (n *featureHigherTFTrendNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *featureHigherTFTrendNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	bars := artifact.MustGet(av, inputKeyMarketOHLCVBarsH1)
	if len(bars) < 2 {
		return fmt.Errorf("feature_higher_tf_trend node %q: market.ohlcv_bars.h1 requires at least 2 bars", n.id)
	}
	first := bars[0]
	last := bars[len(bars)-1]
	direction := "flat"
	switch {
	case last.Close.Gt(first.Open):
		direction = "up"
	case last.Close.Lt(first.Open):
		direction = "down"
	}
	artifact.Set(aw, featureHigherTFTrendOutputKey(n.id), algotrade.FeatureHigherTFTrend{
		Direction: direction,
	})
	return nil
}

func (n *featureSessionStateNode) Name() string { return "dagruntime.feature_session_state." + n.id }
func (n *featureSessionStateNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{usecase.InputKeyMarketOHLCVBars}
}
func (n *featureSessionStateNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{featureSessionStateOutputKey(n.id)}
}
func (n *featureSessionStateNode) Reads() []state.AnyKey  { return nil }
func (n *featureSessionStateNode) Writes() []state.AnyKey { return nil }
func (n *featureSessionStateNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *featureSessionStateNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	bars := artifact.MustGet(av, usecase.InputKeyMarketOHLCVBars)
	if len(bars) == 0 {
		return fmt.Errorf("feature_session_state node %q: market.ohlcv_bars is empty", n.id)
	}
	last := bars[len(bars)-1]
	ts := last.Closetime
	if ts.IsZero() {
		ts = last.Opentime
	}
	if ts.IsZero() {
		return fmt.Errorf("feature_session_state node %q: last bar has no open/close time", n.id)
	}
	artifact.Set(aw, featureSessionStateOutputKey(n.id), algotrade.FeatureSessionState{
		Session: sessionNameByHour(ts.Time().UTC().Hour()),
	})
	return nil
}

type signalBreakoutLongNode struct {
	id          string
	rangeNodeID string
}

func (n *signalBreakoutLongNode) Name() string { return "dagruntime.signal_breakout_long." + n.id }
func (n *signalBreakoutLongNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		usecase.InputKeyMarketOHLCVBars,
		featureRangeHighOutputKey(n.rangeNodeID),
	}
}
func (n *signalBreakoutLongNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{signalBreakoutLongOutputKey(n.id)}
}
func (n *signalBreakoutLongNode) Reads() []state.AnyKey  { return nil }
func (n *signalBreakoutLongNode) Writes() []state.AnyKey { return nil }
func (n *signalBreakoutLongNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}

func (n *signalBreakoutLongNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	bars := artifact.MustGet(av, usecase.InputKeyMarketOHLCVBars)
	if len(bars) == 0 {
		return fmt.Errorf("signal_breakout_long node %q: market.ohlcv_bars is empty", n.id)
	}
	rangeHigh := artifact.MustGet(av, featureRangeHighOutputKey(n.rangeNodeID))
	lastClose := bars[len(bars)-1].Close
	artifact.Set(aw, signalBreakoutLongOutputKey(n.id), algotrade.BreakoutLongCandidate{
		Triggered: lastClose.Gt(rangeHigh.Value),
	})
	return nil
}

type observabilitySignalDecision struct {
	Action       algotrade.OrderAction
	Reason       string
	Allowed      bool
	PositionSize float64
}

type filterSessionNode struct {
	id              string
	signalNodeID    string
	allowedSessions []string
}

type signalExitBasicNode struct {
	id             string
	positionNodeID string
	stopLossNodeID string
}

func (n *signalExitBasicNode) Name() string { return "dagruntime.signal_exit_basic." + n.id }
func (n *signalExitBasicNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		positionSnapshotOutputKey(n.positionNodeID),
		riskStopLossDistanceOutputKey(n.stopLossNodeID),
		usecase.InputKeyMarketOHLCVBars,
	}
}
func (n *signalExitBasicNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{signalExitDecisionOutputKey(n.id)}
}
func (n *signalExitBasicNode) Reads() []state.AnyKey  { return nil }
func (n *signalExitBasicNode) Writes() []state.AnyKey { return nil }
func (n *signalExitBasicNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *signalExitBasicNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	position := artifact.MustGet(av, positionSnapshotOutputKey(n.positionNodeID))
	if !position.HasPosition || position.Side != algotrade.PositionSideLong || position.Size <= 0 {
		artifact.Set(aw, signalExitDecisionOutputKey(n.id), algotrade.ExitDecision{
			ShouldExit: false,
			Reason:     "no_position",
		})
		return nil
	}
	stopLossDistance := artifact.MustGet(av, riskStopLossDistanceOutputKey(n.stopLossNodeID))
	if stopLossDistance.Raw() <= 0 {
		return fmt.Errorf("signal_exit_basic node %q: stop loss distance must be > 0", n.id)
	}
	bars := artifact.MustGet(av, usecase.InputKeyMarketOHLCVBars)
	if len(bars) == 0 {
		return fmt.Errorf("signal_exit_basic node %q: market.ohlcv_bars is empty", n.id)
	}
	lastClose := bars[len(bars)-1].Close
	stopPrice := position.EntryPrice.Sub(stopLossDistance)
	decision := algotrade.ExitDecision{
		ShouldExit: false,
		Reason:     "hold_position",
	}
	if lastClose.Le(stopPrice) {
		decision.ShouldExit = true
		decision.Reason = "stop_loss_hit"
	}
	artifact.Set(aw, signalExitDecisionOutputKey(n.id), decision)
	return nil
}

func (n *filterSessionNode) Name() string { return "dagruntime.filter_session." + n.id }
func (n *filterSessionNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		signalBreakoutLongOutputKey(n.signalNodeID),
		usecase.InputKeyMarketOHLCVBars,
	}
}
func (n *filterSessionNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{entryFilterResultKey(n.id)}
}
func (n *filterSessionNode) Reads() []state.AnyKey    { return nil }
func (n *filterSessionNode) Writes() []state.AnyKey   { return nil }
func (n *filterSessionNode) Spec() node.ExecutionSpec { return node.ExecutionSpec{Deterministic: true} }

func (n *filterSessionNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	candidate := artifact.MustGet(av, signalBreakoutLongOutputKey(n.signalNodeID))
	if !candidate.Triggered {
		artifact.Set(aw, entryFilterResultKey(n.id), algotrade.EntryFilterResult{Allowed: false, Reason: "no_signal"})
		return nil
	}
	bars := artifact.MustGet(av, usecase.InputKeyMarketOHLCVBars)
	if len(bars) == 0 {
		return fmt.Errorf("filter_session node %q: market.ohlcv_bars is empty", n.id)
	}
	last := bars[len(bars)-1]
	ts := last.Closetime
	if ts.IsZero() {
		ts = last.Opentime
	}
	if ts.IsZero() {
		return fmt.Errorf("filter_session node %q: last bar has no open/close time", n.id)
	}
	if !isAllowedSession(ts.Time().UTC(), n.allowedSessions) {
		artifact.Set(aw, entryFilterResultKey(n.id), algotrade.EntryFilterResult{Allowed: false, Reason: "session_blocked"})
		return nil
	}
	artifact.Set(aw, entryFilterResultKey(n.id), algotrade.EntryFilterResult{Allowed: true})
	return nil
}

type filterSpreadNode struct {
	id            string
	allowedNodeID string
	maxSpreadBps  float64
}

func (n *filterSpreadNode) Name() string { return "dagruntime.filter_spread." + n.id }
func (n *filterSpreadNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		entryFilterResultKey(n.allowedNodeID),
		usecase.InputKeyMarketSpreadBps,
	}
}
func (n *filterSpreadNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{entryFilterResultKey(n.id)}
}
func (n *filterSpreadNode) Reads() []state.AnyKey    { return nil }
func (n *filterSpreadNode) Writes() []state.AnyKey   { return nil }
func (n *filterSpreadNode) Spec() node.ExecutionSpec { return node.ExecutionSpec{Deterministic: true} }

func (n *filterSpreadNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	current := artifact.MustGet(av, entryFilterResultKey(n.allowedNodeID))
	if !current.Allowed {
		artifact.Set(aw, entryFilterResultKey(n.id), current)
		return nil
	}
	spread := artifact.MustGet(av, usecase.InputKeyMarketSpreadBps)
	if spread > n.maxSpreadBps {
		artifact.Set(aw, entryFilterResultKey(n.id), algotrade.EntryFilterResult{Allowed: false, Reason: "spread_limit"})
		return nil
	}
	artifact.Set(aw, entryFilterResultKey(n.id), current)
	return nil
}

type filterDailyLossLimitNode struct {
	id            string
	allowedNodeID string
	maxDailyLoss  float64
}

type filterEconomicEventNode struct {
	id            string
	allowedNodeID string
	blockSeverity string
}

func (n *filterEconomicEventNode) Name() string { return "dagruntime.filter_economic_event." + n.id }
func (n *filterEconomicEventNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		entryFilterResultKey(n.allowedNodeID),
		inputKeyEconomicEvents,
	}
}
func (n *filterEconomicEventNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{entryFilterResultKey(n.id)}
}
func (n *filterEconomicEventNode) Reads() []state.AnyKey  { return nil }
func (n *filterEconomicEventNode) Writes() []state.AnyKey { return nil }
func (n *filterEconomicEventNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *filterEconomicEventNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	current := artifact.MustGet(av, entryFilterResultKey(n.allowedNodeID))
	if !current.Allowed {
		artifact.Set(aw, entryFilterResultKey(n.id), current)
		return nil
	}
	events := artifact.MustGet(av, inputKeyEconomicEvents)
	for _, e := range events {
		severity := strings.ToLower(strings.TrimSpace(e.Severity))
		if severity == n.blockSeverity {
			artifact.Set(aw, entryFilterResultKey(n.id), algotrade.EntryFilterResult{
				Allowed: false,
				Reason:  "economic_event",
			})
			return nil
		}
	}
	artifact.Set(aw, entryFilterResultKey(n.id), current)
	return nil
}

type filterHigherTFAlignmentNode struct {
	id                string
	allowedNodeID     string
	trendNodeID       string
	requiredDirection string
}

func (n *filterHigherTFAlignmentNode) Name() string {
	return "dagruntime.filter_higher_tf_alignment." + n.id
}
func (n *filterHigherTFAlignmentNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		entryFilterResultKey(n.allowedNodeID),
		featureHigherTFTrendOutputKey(n.trendNodeID),
	}
}
func (n *filterHigherTFAlignmentNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{entryFilterResultKey(n.id)}
}
func (n *filterHigherTFAlignmentNode) Reads() []state.AnyKey  { return nil }
func (n *filterHigherTFAlignmentNode) Writes() []state.AnyKey { return nil }
func (n *filterHigherTFAlignmentNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *filterHigherTFAlignmentNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	current := artifact.MustGet(av, entryFilterResultKey(n.allowedNodeID))
	if !current.Allowed {
		artifact.Set(aw, entryFilterResultKey(n.id), current)
		return nil
	}
	trend := artifact.MustGet(av, featureHigherTFTrendOutputKey(n.trendNodeID))
	if trend.Direction != n.requiredDirection {
		artifact.Set(aw, entryFilterResultKey(n.id), algotrade.EntryFilterResult{
			Allowed: false,
			Reason:  "higher_tf_alignment",
		})
		return nil
	}
	artifact.Set(aw, entryFilterResultKey(n.id), current)
	return nil
}

func (n *filterDailyLossLimitNode) Name() string { return "dagruntime.filter_daily_loss_limit." + n.id }
func (n *filterDailyLossLimitNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{entryFilterResultKey(n.allowedNodeID)}
}
func (n *filterDailyLossLimitNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{entryFilterResultKey(n.id)}
}
func (n *filterDailyLossLimitNode) Reads() []state.AnyKey {
	return []state.AnyKey{algotrade.StateDailyPnL}
}
func (n *filterDailyLossLimitNode) Writes() []state.AnyKey { return nil }
func (n *filterDailyLossLimitNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *filterDailyLossLimitNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	current := artifact.MustGet(av, entryFilterResultKey(n.allowedNodeID))
	if !current.Allowed {
		artifact.Set(aw, entryFilterResultKey(n.id), current)
		return nil
	}
	daily, ok := state.Get(txn, algotrade.StateDailyPnL)
	if !ok {
		artifact.Set(aw, entryFilterResultKey(n.id), current)
		return nil
	}
	limitHit := daily.LossLimitHit
	if !limitHit && n.maxDailyLoss > 0 && daily.RealizedPnL <= -n.maxDailyLoss {
		limitHit = true
	}
	if limitHit {
		artifact.Set(aw, entryFilterResultKey(n.id), algotrade.EntryFilterResult{
			Allowed: false,
			Reason:  "daily_loss_limit",
		})
		return nil
	}
	artifact.Set(aw, entryFilterResultKey(n.id), current)
	return nil
}

type riskPositionSizingNode struct {
	id              string
	allowedNodeID   string
	atrNodeID       string
	riskRate        float64
	stopATRMultiple float64
}

type riskStopLossFromATRNode struct {
	id          string
	atrNodeID   string
	atrMultiple float64
}

type riskMaxPositionsCheckNode struct {
	id            string
	allowedNodeID string
	maxPositions  int
}

func (n *riskMaxPositionsCheckNode) Name() string {
	return "dagruntime.risk_max_positions_check." + n.id
}
func (n *riskMaxPositionsCheckNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{entryFilterResultKey(n.allowedNodeID)}
}
func (n *riskMaxPositionsCheckNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{entryFilterResultKey(n.id)}
}
func (n *riskMaxPositionsCheckNode) Reads() []state.AnyKey {
	return []state.AnyKey{algotrade.StateOpenPositions}
}
func (n *riskMaxPositionsCheckNode) Writes() []state.AnyKey { return nil }
func (n *riskMaxPositionsCheckNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *riskMaxPositionsCheckNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	current := artifact.MustGet(av, entryFilterResultKey(n.allowedNodeID))
	if !current.Allowed {
		artifact.Set(aw, entryFilterResultKey(n.id), current)
		return nil
	}
	openPositions, ok := state.Get(txn, algotrade.StateOpenPositions)
	if !ok {
		artifact.Set(aw, entryFilterResultKey(n.id), current)
		return nil
	}
	if len(openPositions.Items) >= n.maxPositions {
		artifact.Set(aw, entryFilterResultKey(n.id), algotrade.EntryFilterResult{
			Allowed: false,
			Reason:  "max_positions",
		})
		return nil
	}
	artifact.Set(aw, entryFilterResultKey(n.id), current)
	return nil
}

func (n *riskStopLossFromATRNode) Name() string { return "dagruntime.risk_stop_loss_from_atr." + n.id }
func (n *riskStopLossFromATRNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{featureATROutputKey(n.atrNodeID)}
}
func (n *riskStopLossFromATRNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{riskStopLossDistanceOutputKey(n.id)}
}
func (n *riskStopLossFromATRNode) Reads() []state.AnyKey  { return nil }
func (n *riskStopLossFromATRNode) Writes() []state.AnyKey { return nil }
func (n *riskStopLossFromATRNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *riskStopLossFromATRNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	atr := artifact.MustGet(av, featureATROutputKey(n.atrNodeID))
	if atr.Value.Raw() <= 0 {
		return fmt.Errorf("risk_stop_loss_from_atr node %q: atr must be > 0", n.id)
	}
	stopRaw := int64(float64(atr.Value.Raw()) * n.atrMultiple)
	if stopRaw <= 0 {
		return fmt.Errorf("risk_stop_loss_from_atr node %q: computed stop distance must be > 0", n.id)
	}
	artifact.Set(aw, riskStopLossDistanceOutputKey(n.id), marketdata.NewPriceFromRaw(stopRaw))
	return nil
}

type riskTakeProfitFromRRNode struct {
	id             string
	stopLossNodeID string
	rrRatio        float64
}

func (n *riskTakeProfitFromRRNode) Name() string {
	return "dagruntime.risk_take_profit_from_rr." + n.id
}
func (n *riskTakeProfitFromRRNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{riskStopLossDistanceOutputKey(n.stopLossNodeID)}
}
func (n *riskTakeProfitFromRRNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{riskTakeProfitDistanceOutputKey(n.id)}
}
func (n *riskTakeProfitFromRRNode) Reads() []state.AnyKey  { return nil }
func (n *riskTakeProfitFromRRNode) Writes() []state.AnyKey { return nil }
func (n *riskTakeProfitFromRRNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *riskTakeProfitFromRRNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	stopLossDistance := artifact.MustGet(av, riskStopLossDistanceOutputKey(n.stopLossNodeID))
	if stopLossDistance.Raw() <= 0 {
		return fmt.Errorf("risk_take_profit_from_rr node %q: stop loss distance must be > 0", n.id)
	}
	takeProfitRaw := int64(float64(stopLossDistance.Raw()) * n.rrRatio)
	if takeProfitRaw <= 0 {
		return fmt.Errorf("risk_take_profit_from_rr node %q: computed take profit distance must be > 0", n.id)
	}
	artifact.Set(aw, riskTakeProfitDistanceOutputKey(n.id), marketdata.NewPriceFromRaw(takeProfitRaw))
	return nil
}

func (n *riskPositionSizingNode) Name() string { return "dagruntime.risk_position_sizing." + n.id }
func (n *riskPositionSizingNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		entryFilterResultKey(n.allowedNodeID),
		featureATROutputKey(n.atrNodeID),
		usecase.InputKeyAccountBalance,
	}
}
func (n *riskPositionSizingNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{
		riskStopLossDistanceOutputKey(n.id),
		riskPositionSizeOutputKey(n.id),
	}
}
func (n *riskPositionSizingNode) Reads() []state.AnyKey  { return nil }
func (n *riskPositionSizingNode) Writes() []state.AnyKey { return nil }
func (n *riskPositionSizingNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}

func (n *riskPositionSizingNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	filtered := artifact.MustGet(av, entryFilterResultKey(n.allowedNodeID))
	if !filtered.Allowed {
		artifact.Set(aw, riskStopLossDistanceOutputKey(n.id), marketdata.Price(0))
		artifact.Set(aw, riskPositionSizeOutputKey(n.id), 0.0)
		return nil
	}
	atr := artifact.MustGet(av, featureATROutputKey(n.atrNodeID))
	if atr.Value.Raw() <= 0 {
		return fmt.Errorf("risk_position_sizing node %q: atr must be > 0", n.id)
	}
	stopRaw := int64(float64(atr.Value.Raw()) * n.stopATRMultiple)
	if stopRaw <= 0 {
		return fmt.Errorf("risk_position_sizing node %q: computed stop distance must be > 0", n.id)
	}
	accountBalance := artifact.MustGet(av, usecase.InputKeyAccountBalance)
	if accountBalance <= 0 {
		return fmt.Errorf("risk_position_sizing node %q: account.balance must be > 0", n.id)
	}
	riskAmount := accountBalance * n.riskRate
	positionSize := riskAmount / float64(stopRaw)
	artifact.Set(aw, riskStopLossDistanceOutputKey(n.id), marketdata.NewPriceFromRaw(stopRaw))
	artifact.Set(aw, riskPositionSizeOutputKey(n.id), positionSize)
	return nil
}

type signalDecisionMapperNode struct {
	id            string
	allowedNodeID string
	sizingNodeID  string
}

func (n *signalDecisionMapperNode) Name() string { return "dagruntime.signal_decision_mapper." + n.id }
func (n *signalDecisionMapperNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		entryFilterResultKey(n.allowedNodeID),
		riskStopLossDistanceOutputKey(n.sizingNodeID),
		riskPositionSizeOutputKey(n.sizingNodeID),
	}
}
func (n *signalDecisionMapperNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{signalDecisionOutputKey(n.id)}
}
func (n *signalDecisionMapperNode) Reads() []state.AnyKey  { return nil }
func (n *signalDecisionMapperNode) Writes() []state.AnyKey { return nil }
func (n *signalDecisionMapperNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}

func (n *signalDecisionMapperNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	filtered := artifact.MustGet(av, entryFilterResultKey(n.allowedNodeID))
	stopLossDistance := artifact.MustGet(av, riskStopLossDistanceOutputKey(n.sizingNodeID))
	positionSize := artifact.MustGet(av, riskPositionSizeOutputKey(n.sizingNodeID))
	decision := algotrade.OrderRequest{
		Action:           algotrade.OrderActionHold,
		Reason:           filtered.Reason,
		PositionSize:     0,
		StopLossDistance: marketdata.Price(0),
	}
	if filtered.Allowed && positionSize > 0 && stopLossDistance.Raw() > 0 {
		decision.Action = algotrade.OrderActionBuy
		decision.Reason = "entry_allowed"
		decision.PositionSize = positionSize
		decision.StopLossDistance = stopLossDistance
	}
	artifact.Set(aw, signalDecisionOutputKey(n.id), decision)
	return nil
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

func featureATROutputKey(nodeID string) artifact.Key[algotrade.FeatureATR] {
	return artifact.Key[algotrade.FeatureATR]{
		Name:     fmt.Sprintf("%s.atr", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.feature_atr.%s.v1", nodeID),
	}
}

func featureRangeHighOutputKey(nodeID string) artifact.Key[algotrade.FeatureRangeHigh] {
	return artifact.Key[algotrade.FeatureRangeHigh]{
		Name:     fmt.Sprintf("%s.range_high", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.feature_range_high.%s.v1", nodeID),
	}
}

func featureRangeLowOutputKey(nodeID string) artifact.Key[algotrade.FeatureRangeLow] {
	return artifact.Key[algotrade.FeatureRangeLow]{
		Name:     fmt.Sprintf("%s.range_low", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.feature_range_low.%s.v1", nodeID),
	}
}

func featureSpreadOutputKey(nodeID string) artifact.Key[algotrade.FeatureSpread] {
	return artifact.Key[algotrade.FeatureSpread]{
		Name:     fmt.Sprintf("%s.spread", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.feature_spread.%s.v1", nodeID),
	}
}

func featureSessionStateOutputKey(nodeID string) artifact.Key[algotrade.FeatureSessionState] {
	return artifact.Key[algotrade.FeatureSessionState]{
		Name:     fmt.Sprintf("%s.session_state", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.feature_session_state.%s.v1", nodeID),
	}
}

func featureHigherTFTrendOutputKey(nodeID string) artifact.Key[algotrade.FeatureHigherTFTrend] {
	return artifact.Key[algotrade.FeatureHigherTFTrend]{
		Name:     fmt.Sprintf("%s.higher_tf_trend", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.feature_higher_tf_trend.%s.v1", nodeID),
	}
}

var inputKeyEconomicEvents = artifact.Key[[]algotrade.EconomicEvent]{
	Name:     "market.economic_events",
	StableID: "artifact:input.market.economic_events.v1",
}

var inputKeyMarketOHLCVBarsH1 = artifact.Key[[]marketdata.OHLCV]{
	Name:     "market.ohlcv_bars.h1",
	StableID: "artifact:input.market.ohlcv_bars.h1.v1",
}

func signalBreakoutLongOutputKey(nodeID string) artifact.Key[algotrade.BreakoutLongCandidate] {
	return artifact.Key[algotrade.BreakoutLongCandidate]{
		Name:     fmt.Sprintf("%s.breakout_long", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.signal_breakout_long.%s.v1", nodeID),
	}
}

func signalExitDecisionOutputKey(nodeID string) artifact.Key[algotrade.ExitDecision] {
	return artifact.Key[algotrade.ExitDecision]{
		Name:     fmt.Sprintf("%s.exit_decision", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.signal_exit_basic.%s.v1", nodeID),
	}
}

func entryFilterResultKey(nodeID string) artifact.Key[algotrade.EntryFilterResult] {
	return artifact.Key[algotrade.EntryFilterResult]{
		Name:     fmt.Sprintf("%s.entry_filter_result", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.filter.entry.%s.v1", nodeID),
	}
}

func riskStopLossDistanceOutputKey(nodeID string) artifact.Key[marketdata.Price] {
	return artifact.Key[marketdata.Price]{
		Name:     fmt.Sprintf("%s.stop_loss_distance", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.risk.stop_loss_distance.%s.v1", nodeID),
	}
}

func riskPositionSizeOutputKey(nodeID string) artifact.Key[float64] {
	return artifact.Key[float64]{
		Name:     fmt.Sprintf("%s.position_size", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.risk.position_size.%s.v1", nodeID),
	}
}

func riskTakeProfitDistanceOutputKey(nodeID string) artifact.Key[marketdata.Price] {
	return artifact.Key[marketdata.Price]{
		Name:     fmt.Sprintf("%s.take_profit_distance", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.risk.take_profit_distance.%s.v1", nodeID),
	}
}

func signalDecisionOutputKey(nodeID string) artifact.Key[algotrade.OrderRequest] {
	return artifact.Key[algotrade.OrderRequest]{
		Name:     fmt.Sprintf("%s.signal_order_decision", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.signal.decision.%s.v1", nodeID),
	}
}

func observabilitySignalDecisionOutputKey(nodeID string) artifact.Key[observabilitySignalDecision] {
	return artifact.Key[observabilitySignalDecision]{
		Name:     fmt.Sprintf("%s.observability_signal_decision", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.observability.signal_decision.%s.v1", nodeID),
	}
}

func paperExecutionResultOutputKey(nodeID string) artifact.Key[algotrade.PaperExecutionResult] {
	return artifact.Key[algotrade.PaperExecutionResult]{
		Name:     fmt.Sprintf("%s.paper_execution_result", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.execution.paper.%s.v1", nodeID),
	}
}

func positionSnapshotOutputKey(nodeID string) artifact.Key[algotrade.PositionSnapshot] {
	return artifact.Key[algotrade.PositionSnapshot]{
		Name:     fmt.Sprintf("%s.position_snapshot", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.position.snapshot.%s.v1", nodeID),
	}
}

func requiredFloat(config map[string]any, key string) (float64, error) {
	if config == nil {
		return 0, fmt.Errorf("config.%s is required", key)
	}
	raw, ok := config[key]
	if !ok {
		return 0, fmt.Errorf("config.%s is required", key)
	}
	return parseFloatConfig(raw, key)
}

func optionalFloat(config map[string]any, key string, defaultValue float64) float64 {
	if config == nil {
		return defaultValue
	}
	raw, ok := config[key]
	if !ok {
		return defaultValue
	}
	value, err := parseFloatConfig(raw, key)
	if err != nil {
		return defaultValue
	}
	return value
}

func parseFloatConfig(raw any, key string) (float64, error) {
	switch v := raw.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("config.%s must be number", key)
	}
}

func optionalStringSlice(config map[string]any, key string, defaultValue []string) []string {
	if config == nil {
		return defaultValue
	}
	raw, ok := config[key]
	if !ok {
		return defaultValue
	}
	switch v := raw.(type) {
	case []string:
		return normalizeSessions(v)
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return defaultValue
			}
			out = append(out, s)
		}
		return normalizeSessions(out)
	default:
		return defaultValue
	}
}

func normalizeSessions(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		trimmed := strings.TrimSpace(strings.ToLower(s))
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func isAllowedSession(t time.Time, sessions []string) bool {
	hour := t.Hour()
	for _, s := range sessions {
		switch s {
		case "tokyo":
			if hour >= 0 && hour < 9 {
				return true
			}
		case "london":
			if hour >= 7 && hour < 16 {
				return true
			}
		case "newyork":
			if hour >= 13 && hour < 22 {
				return true
			}
		}
	}
	return false
}

func sessionNameByHour(hour int) string {
	switch {
	case hour >= 0 && hour < 9:
		return "tokyo"
	case hour >= 7 && hour < 16:
		return "london"
	case hour >= 13 && hour < 22:
		return "newyork"
	default:
		return "off"
	}
}
