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

// TradeIntent is the normalized strategy intent produced by entry decision workflows.
type TradeIntent struct {
	IntentID        string
	Symbol          string
	Action          OrderAction
	Reason          string
	PositionSize    float64
	InitialStopLoss marketdata.Price

	StrategyID      string
	WorkflowName    string
	WorkflowVersion string
	ParameterSetID  string

	CreatedAt marketdata.UTCTime
}

// OrderRequest is the order intent derived from strategy decision nodes.
// It remains as a workflow-local compatibility type and should gradually converge
// toward TradeIntent or an execution-local DTO depending on the caller boundary.
type OrderRequest struct {
	Action           OrderAction
	Reason           string
	PositionSize     float64
	StopLossDistance marketdata.Price
}

func (r OrderRequest) ToTradeIntent(intentID, symbol, workflowName, workflowVersion string, createdAt marketdata.UTCTime) TradeIntent {
	return TradeIntent{
		IntentID:        intentID,
		Symbol:          symbol,
		Action:          r.Action,
		Reason:          r.Reason,
		PositionSize:    r.PositionSize,
		InitialStopLoss: r.StopLossDistance,
		WorkflowName:    workflowName,
		WorkflowVersion: workflowVersion,
		CreatedAt:       createdAt,
	}
}

func (i TradeIntent) ToOrderRequest() OrderRequest {
	return OrderRequest{
		Action:           i.Action,
		Reason:           i.Reason,
		PositionSize:     i.PositionSize,
		StopLossDistance: i.InitialStopLoss,
	}
}

type ExecutionStatus string

const (
	ExecutionStatusRequested       ExecutionStatus = "requested"
	ExecutionStatusSubmitted       ExecutionStatus = "submitted"
	ExecutionStatusAccepted        ExecutionStatus = "accepted"
	ExecutionStatusPartiallyFilled ExecutionStatus = "partially_filled"
	ExecutionStatusFilled          ExecutionStatus = "filled"
	ExecutionStatusRejected        ExecutionStatus = "rejected"
	ExecutionStatusCancelled       ExecutionStatus = "cancelled"
)

// ExecutionResult is the normalized execution progression state derived from a trade intent.
type ExecutionResult struct {
	ExecutionID string
	IntentID    string
	OrderID     string
	Status      ExecutionStatus

	Action        OrderAction
	RequestedSize float64
	FilledSize    float64

	Reason   string
	SourceID string

	RequestedAt marketdata.UTCTime
	UpdatedAt   marketdata.UTCTime
}

// PaperExecutionResult represents the simulated execution result in v0.
type PaperExecutionResult struct {
	Submitted bool
	Action    OrderAction
	Size      float64
	Reason    string
}

// MarketOrderRequest represents a broker-facing market order submission request/result.
// It remains as an execution-edge compatibility type and can be projected into
// ExecutionResult when the execution workflow switches to the new primary model.
type MarketOrderRequest struct {
	Submitted bool
	OrderID   string
	Action    OrderAction
	Size      float64
	Reason    string
	SourceID  string
}

func (r MarketOrderRequest) ToExecutionResult(executionID, intentID string, at marketdata.UTCTime) ExecutionResult {
	status := ExecutionStatusRequested
	if r.Submitted {
		status = ExecutionStatusSubmitted
	}
	return ExecutionResult{
		ExecutionID:   executionID,
		IntentID:      intentID,
		OrderID:       r.OrderID,
		Status:        status,
		Action:        r.Action,
		RequestedSize: r.Size,
		Reason:        r.Reason,
		SourceID:      r.SourceID,
		RequestedAt:   at,
		UpdatedAt:     at,
	}
}

// FillResult represents the normalized fill confirmation outcome.
// It remains as a fill-edge compatibility type and can be projected into
// ExecutionResult while the execution nodes are migrated incrementally.
type FillResult struct {
	Filled  bool
	OrderID string
	Action  OrderAction
	Size    float64
	Reason  string
}

func (r FillResult) ToExecutionResult(executionID, intentID string, requestedAt, updatedAt marketdata.UTCTime) ExecutionResult {
	status := ExecutionStatusAccepted
	if r.Filled {
		status = ExecutionStatusFilled
	}
	return ExecutionResult{
		ExecutionID:   executionID,
		IntentID:      intentID,
		OrderID:       r.OrderID,
		Status:        status,
		Action:        r.Action,
		RequestedSize: r.Size,
		FilledSize:    r.Size,
		Reason:        r.Reason,
		RequestedAt:   requestedAt,
		UpdatedAt:     updatedAt,
	}
}

type PositionSide string

const (
	PositionSideFlat  PositionSide = "flat"
	PositionSideLong  PositionSide = "long"
	PositionSideShort PositionSide = "short"
)

// PositionSnapshot is a lightweight artifact/view representation of an open position.
type PositionSnapshot struct {
	HasPosition bool
	Side        PositionSide
	Size        float64
	EntryPrice  marketdata.Price
}

// ClosedTrade is the minimal completed-trade fact used for hypothesis validation and summary projection.
type ClosedTrade struct {
	TradeID    string
	IntentID   string
	PositionID string

	Symbol string
	Side   PositionSide
	Size   float64

	EntryTime  marketdata.UTCTime
	ExitTime   marketdata.UTCTime
	EntryPrice marketdata.Price
	ExitPrice  marketdata.Price

	GrossPnL        float64
	NetPnL          float64
	ExitReason      string
	HoldingDuration string

	StrategyID      string
	WorkflowName    string
	WorkflowVersion string
	ParameterSetID  string
}

// PnLSnapshot is the normalized pnl view used as a bridge between closed-trade facts and summaries.
type PnLSnapshot struct {
	TradingDay    string
	RealizedPnL   float64
	UnrealizedPnL float64
	PeakEquity    float64
	Drawdown      float64
	LossLimitHit  bool
}

// StrategySummary is the minimal projection used to compare strategy outcomes.
type StrategySummary struct {
	StrategyID      string
	WorkflowName    string
	WorkflowVersion string
	ParameterSetID  string

	TradeCount   int
	WinCount     int
	LossCount    int
	WinRate      float64
	TotalNetPnL  float64
	AverageWin   float64
	AverageLoss  float64
	ProfitFactor float64
	MaxDrawdown  float64

	UpdatedAt marketdata.UTCTime
}
