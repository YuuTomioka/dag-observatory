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

// BreakoutLongCandidate is the normalized signal candidate emitted by breakout nodes.
type BreakoutLongCandidate struct {
	Triggered bool
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
