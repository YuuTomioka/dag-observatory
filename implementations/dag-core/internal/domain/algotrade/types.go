package algotrade

import "dag-observatory/dag-core/internal/domain/marketdata"

type MarketSnapshot struct {
	Symbol    string
	SpreadBps float64
}

type FeatureATR struct {
	Value marketdata.Price
}

type FeatureRangeHigh struct {
	Value marketdata.Price
}

type FeatureRangeLow struct {
	Value marketdata.Price
}

type FeatureSpread struct {
	Bps float64
}

type FeatureSessionState struct {
	Session string
}

type FeatureHigherTFTrend struct {
	Direction string
}

type EconomicEvent struct {
	Name     string
	Severity string
	StartsAt string
	EndsAt   string
}

// BreakoutLongCandidate is the normalized signal candidate emitted by breakout nodes.
type BreakoutLongCandidate struct {
	Triggered bool
}

// BreakoutShortCandidate is the normalized short-side signal candidate emitted by breakout nodes.
type BreakoutShortCandidate struct {
	Triggered bool
}

type ExitDecision struct {
	ShouldExit bool
	Reason     string
}

type EntryFilterResult struct {
	Allowed bool
	Reason  string
}

type OrderAction string

const (
	OrderActionHold OrderAction = "hold"
	OrderActionBuy  OrderAction = "buy"
	OrderActionSell OrderAction = "sell"
)

// OrderRequest is the order intent derived from strategy decision nodes.
type OrderRequest struct {
	Action           OrderAction
	Reason           string
	PositionSize     float64
	StopLossDistance marketdata.Price
}

// PaperExecutionResult represents the simulated execution result in v0.
type PaperExecutionResult struct {
	Submitted bool
	Action    OrderAction
	Size      float64
	Reason    string
}

// MarketOrderRequest represents a broker-facing market order submission request/result.
type MarketOrderRequest struct {
	Submitted bool
	OrderID   string
	Action    OrderAction
	Size      float64
	Reason    string
	SourceID  string
}

// FillResult represents the normalized fill confirmation outcome.
type FillResult struct {
	Filled  bool
	OrderID string
	Action  OrderAction
	Size    float64
	Reason  string
}

type PositionSide string

const (
	PositionSideFlat  PositionSide = "flat"
	PositionSideLong  PositionSide = "long"
	PositionSideShort PositionSide = "short"
)

type PositionSnapshot struct {
	HasPosition bool
	Side        PositionSide
	Size        float64
	EntryPrice  marketdata.Price
}
