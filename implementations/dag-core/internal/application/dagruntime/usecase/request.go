package usecase

import "dag-observatory/dag-core/internal/domain/marketdata"

type MarketdataRunInput struct {
	SymbolID      int64
	SymbolCode    string
	TimeframeCode string
	From          marketdata.UTCTime
	To            marketdata.UTCTime
}

type RunWorkflowRequest struct {
	Symbol          string
	Mode            string
	Partition       string
	Bars            []float64
	OHLCVBars       []marketdata.OHLCV
	SpreadBps       float64
	FeeBps          float64
	SlippageBps     float64
	MinLot          float64
	WorkflowName    string
	WorkflowVersion string
	ParameterSetID  string
	AccountBalance  float64
	Marketdata      *MarketdataRunInput
	RunID           string
}

type RunWorkflowResult struct {
	RunID                string
	Symbol               string
	Mode                 string
	Marketdata           *MarketdataRunInput
	TaskID               string
	TaskName             string
	Attempt              int
	EnqueueMode          bool
	MessagingDestination string
}
