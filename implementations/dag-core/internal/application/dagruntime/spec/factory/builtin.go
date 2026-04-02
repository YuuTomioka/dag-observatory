package factory

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strings"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	marketdatarepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type Dependencies struct {
	MarketDataUnitOfWork marketdatarepository.UnitOfWork
}

func NewBuiltinRegistry() (*Registry, error) {
	return NewBuiltinRegistryWithDependencies(Dependencies{})
}

func NewBuiltinRegistryWithDependencies(deps Dependencies) (*Registry, error) {
	registry := NewRegistry()
	factories := []NodeFactory{
		&HeavyCalcFactory{},
		&SMAFactory{},
		&CrossDetectorFactory{},
		&SignalMapperFactory{},
		&MarketTickInputFactory{},
		&MarketBarInputM1Factory{},
		&MarketBarInputH1Factory{},
		&FeatureATRFactory{},
		&FeatureRangeHighFactory{},
		&FeatureRangeLowFactory{},
		&FeatureSpreadFactory{},
		&FeatureSessionStateFactory{},
		&FeatureHigherTFTrendFactory{},
		&SignalBreakoutLongFactory{},
		&SignalExitBasicFactory{},
		&FilterSessionFactory{},
		&FilterSpreadFactory{},
		&FilterEconomicEventFactory{},
		&FilterHigherTFAlignmentFactory{},
		&FilterDailyLossLimitFactory{},
		&RiskPositionSizingFactory{},
		&RiskMaxPositionsCheckFactory{},
		&RiskStopLossFromATRFactory{},
		&RiskTakeProfitFromRRFactory{},
		&SignalDecisionMapperFactory{},
		&ObservabilityEmitSignalDecisionFactory{},
		&ObservabilityEmitOrderDecisionFactory{},
		&ObservabilityEmitPositionEventFactory{},
		&ExecutionSubmitPaperOrderFactory{},
		&ExecutionSubmitMarketOrderFactory{},
		&ExecutionConfirmFillFactory{},
		&PositionSnapshotLoadFactory{},
		&PositionBreakevenFactory{},
		&PositionTrailingStopFactory{},
		&PositionTimeoutExitFactory{},
		&PositionTrackerUpdateFactory{},
	}
	if deps.MarketDataUnitOfWork != nil {
		factories = append(factories,
			&TimeframeBarBackfillFactory{UnitOfWork: deps.MarketDataUnitOfWork},
			&ResolveSymbolFactory{UnitOfWork: deps.MarketDataUnitOfWork},
			&PlanWindowsFactory{},
			&LoadTicksForChunkFactory{UnitOfWork: deps.MarketDataUnitOfWork},
			&AggregateTimeframeBarsFactory{},
			&PersistTimeframeBarsFactory{UnitOfWork: deps.MarketDataUnitOfWork},
		)
	}
	for _, f := range factories {
		if err := registry.Register(f); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

type HeavyCalcFactory struct{}

func (f *HeavyCalcFactory) Kind() string { return "heavy_calc" }

func (f *HeavyCalcFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, nil); err != nil {
		return nil, fmt.Errorf("heavy_calc: %w", err)
	}
	return &heavyCalcNode{}, nil
}

type SMAFactory struct{}

func (f *SMAFactory) Kind() string { return "sma" }

func (f *SMAFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"window", "source"}); err != nil {
		return nil, fmt.Errorf("sma node %q: %w", nodeSpec.ID, err)
	}
	window, err := requiredInt(nodeSpec.Config, "window")
	if err != nil {
		return nil, fmt.Errorf("sma node %q: %w", nodeSpec.ID, err)
	}
	if window <= 0 {
		return nil, fmt.Errorf("sma node %q: config.window must be > 0", nodeSpec.ID)
	}
	source := optionalString(nodeSpec.Config, "source", "close")
	if strings.TrimSpace(source) == "" {
		return nil, fmt.Errorf("sma node %q: config.source must not be empty", nodeSpec.ID)
	}
	return &smaNode{
		id:     nodeSpec.ID,
		window: window,
		source: source,
	}, nil
}

type CrossDetectorFactory struct{}

func (f *CrossDetectorFactory) Kind() string { return "cross_detector" }

func (f *CrossDetectorFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"fast_node_id", "slow_node_id", "mode"}); err != nil {
		return nil, fmt.Errorf("cross_detector node %q: %w", nodeSpec.ID, err)
	}
	fastNodeID, err := requiredString(nodeSpec.Config, "fast_node_id")
	if err != nil {
		return nil, fmt.Errorf("cross_detector node %q: %w", nodeSpec.ID, err)
	}
	slowNodeID, err := requiredString(nodeSpec.Config, "slow_node_id")
	if err != nil {
		return nil, fmt.Errorf("cross_detector node %q: %w", nodeSpec.ID, err)
	}
	mode := optionalString(nodeSpec.Config, "mode", "both")
	if mode != "cross_up" && mode != "cross_down" && mode != "both" {
		return nil, fmt.Errorf("cross_detector node %q: config.mode must be one of cross_up/cross_down/both", nodeSpec.ID)
	}
	return &crossDetectorNode{
		id:         nodeSpec.ID,
		fastNodeID: fastNodeID,
		slowNodeID: slowNodeID,
		mode:       mode,
	}, nil
}

type SignalMapperFactory struct{}

func (f *SignalMapperFactory) Kind() string { return "signal_mapper" }

func (f *SignalMapperFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"cross_node_id", "on_cross_up", "on_cross_down", "on_no_cross"}); err != nil {
		return nil, fmt.Errorf("signal_mapper node %q: %w", nodeSpec.ID, err)
	}
	crossNodeID, err := requiredString(nodeSpec.Config, "cross_node_id")
	if err != nil {
		return nil, fmt.Errorf("signal_mapper node %q: %w", nodeSpec.ID, err)
	}
	return &signalMapperNode{
		id:          nodeSpec.ID,
		crossNodeID: crossNodeID,
		onCrossUp:   optionalString(nodeSpec.Config, "on_cross_up", "buy"),
		onCrossDown: optionalString(nodeSpec.Config, "on_cross_down", "flat"),
		onNoCross:   optionalString(nodeSpec.Config, "on_no_cross", "hold"),
	}, nil
}

type heavyCalcNode struct{}

func (n *heavyCalcNode) Name() string { return "dagruntime.heavy_calc" }

func (n *heavyCalcNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{usecase.InputKeySymbol, usecase.InputKeyMode}
}

func (n *heavyCalcNode) Provides() []artifact.AnyKey { return []artifact.AnyKey{heavyCalcResultKey} }
func (n *heavyCalcNode) Reads() []state.AnyKey       { return nil }
func (n *heavyCalcNode) Writes() []state.AnyKey      { return nil }
func (n *heavyCalcNode) Spec() node.ExecutionSpec    { return node.ExecutionSpec{Deterministic: true} }

func (n *heavyCalcNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	symbol := artifact.MustGet(av, usecase.InputKeySymbol)
	mode := artifact.MustGet(av, usecase.InputKeyMode)
	artifact.Set(aw, heavyCalcResultKey, fmt.Sprintf("%s:%s", symbol, mode))
	return nil
}

var heavyCalcResultKey = artifact.Key[string]{
	Name:     "heavy_calc_result",
	StableID: "artifact:dagruntime.heavy_calc.result.v1",
}

type smaNode struct {
	id     string
	window int
	source string
}

func (n *smaNode) Name() string { return "dagruntime.sma." + n.id }
func (n *smaNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{usecase.InputKeyMarketBars}
}
func (n *smaNode) Provides() []artifact.AnyKey { return []artifact.AnyKey{smaOutputKey(n.id)} }
func (n *smaNode) Reads() []state.AnyKey       { return nil }
func (n *smaNode) Writes() []state.AnyKey      { return nil }
func (n *smaNode) Spec() node.ExecutionSpec    { return node.ExecutionSpec{Deterministic: true} }

func (n *smaNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	_ = n.source
	bars := artifact.MustGet(av, usecase.InputKeyMarketBars)
	if len(bars) == 0 {
		return fmt.Errorf("sma node %q: market.bars is empty", n.id)
	}
	count := n.window
	if count > len(bars) {
		count = len(bars)
	}
	sum := 0.0
	for _, v := range bars[len(bars)-count:] {
		sum += v
	}
	artifact.Set(aw, smaOutputKey(n.id), sum/float64(count))
	return nil
}

type crossDetectorNode struct {
	id         string
	fastNodeID string
	slowNodeID string
	mode       string
}

func (n *crossDetectorNode) Name() string { return "dagruntime.cross_detector." + n.id }
func (n *crossDetectorNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{smaOutputKey(n.fastNodeID), smaOutputKey(n.slowNodeID)}
}
func (n *crossDetectorNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{crossOutputKey(n.id)}
}
func (n *crossDetectorNode) Reads() []state.AnyKey    { return nil }
func (n *crossDetectorNode) Writes() []state.AnyKey   { return nil }
func (n *crossDetectorNode) Spec() node.ExecutionSpec { return node.ExecutionSpec{Deterministic: true} }

func (n *crossDetectorNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	fast := artifact.MustGet(av, smaOutputKey(n.fastNodeID))
	slow := artifact.MustGet(av, smaOutputKey(n.slowNodeID))

	cross := "none"
	switch n.mode {
	case "cross_up":
		if fast > slow {
			cross = "up"
		}
	case "cross_down":
		if fast < slow {
			cross = "down"
		}
	default:
		if fast > slow {
			cross = "up"
		} else if fast < slow {
			cross = "down"
		}
	}
	artifact.Set(aw, crossOutputKey(n.id), cross)
	return nil
}

type signalMapperNode struct {
	id          string
	crossNodeID string
	onCrossUp   string
	onCrossDown string
	onNoCross   string
}

func (n *signalMapperNode) Name() string { return "dagruntime.signal_mapper." + n.id }
func (n *signalMapperNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{crossOutputKey(n.crossNodeID)}
}
func (n *signalMapperNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{signalOutputKey(n.id)}
}
func (n *signalMapperNode) Reads() []state.AnyKey    { return nil }
func (n *signalMapperNode) Writes() []state.AnyKey   { return nil }
func (n *signalMapperNode) Spec() node.ExecutionSpec { return node.ExecutionSpec{Deterministic: true} }

func (n *signalMapperNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	cross := artifact.MustGet(av, crossOutputKey(n.crossNodeID))
	switch cross {
	case "up":
		artifact.Set(aw, signalOutputKey(n.id), n.onCrossUp)
	case "down":
		artifact.Set(aw, signalOutputKey(n.id), n.onCrossDown)
	case "none":
		artifact.Set(aw, signalOutputKey(n.id), n.onNoCross)
	default:
		return fmt.Errorf("signal_mapper node %q: unknown cross value %q", n.id, cross)
	}
	return nil
}

func smaOutputKey(nodeID string) artifact.Key[float64] {
	return artifact.Key[float64]{
		Name:     fmt.Sprintf("%s.sma", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.sma.%s.v1", nodeID),
	}
}

func crossOutputKey(nodeID string) artifact.Key[string] {
	return artifact.Key[string]{
		Name:     fmt.Sprintf("%s.cross", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.cross.%s.v1", nodeID),
	}
}

func signalOutputKey(nodeID string) artifact.Key[string] {
	return artifact.Key[string]{
		Name:     fmt.Sprintf("%s.signal", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.signal.%s.v1", nodeID),
	}
}

func ensureNoUnknownConfigKeys(config map[string]any, allowed []string) error {
	if len(config) == 0 {
		return nil
	}
	for key := range config {
		if !slices.Contains(allowed, key) {
			return fmt.Errorf("config.%s is not supported", key)
		}
	}
	return nil
}

func requiredString(config map[string]any, key string) (string, error) {
	if config == nil {
		return "", fmt.Errorf("config.%s is required", key)
	}
	raw, ok := config[key]
	if !ok {
		return "", fmt.Errorf("config.%s is required", key)
	}
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("config.%s must be string", key)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("config.%s must not be empty", key)
	}
	return value, nil
}

func optionalString(config map[string]any, key, defaultValue string) string {
	if config == nil {
		return defaultValue
	}
	raw, ok := config[key]
	if !ok {
		return defaultValue
	}
	value, ok := raw.(string)
	if !ok {
		return defaultValue
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue
	}
	return value
}

func requiredInt(config map[string]any, key string) (int, error) {
	if config == nil {
		return 0, fmt.Errorf("config.%s is required", key)
	}
	raw, ok := config[key]
	if !ok {
		return 0, fmt.Errorf("config.%s is required", key)
	}
	switch v := raw.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		if math.Trunc(v) != v {
			return 0, fmt.Errorf("config.%s must be integer", key)
		}
		return int(v), nil
	default:
		return 0, fmt.Errorf("config.%s must be integer", key)
	}
}

func optionalInt(config map[string]any, key string, defaultValue int) int {
	if config == nil {
		return defaultValue
	}
	raw, ok := config[key]
	if !ok {
		return defaultValue
	}
	switch v := raw.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		if math.Trunc(v) != v {
			return defaultValue
		}
		return int(v)
	default:
		return defaultValue
	}
}
