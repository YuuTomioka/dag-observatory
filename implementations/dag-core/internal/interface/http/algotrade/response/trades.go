package response

type TradeItem struct {
	TradeID         string  `json:"trade_id"`
	IntentID        string  `json:"intent_id"`
	PositionID      string  `json:"position_id"`
	Symbol          string  `json:"symbol"`
	Side            string  `json:"side"`
	Size            float64 `json:"size"`
	EntryTime       string  `json:"entry_time"`
	ExitTime        string  `json:"exit_time"`
	EntryPriceRaw   int64   `json:"entry_price_raw"`
	ExitPriceRaw    int64   `json:"exit_price_raw"`
	NetPnL          float64 `json:"net_pnl"`
	ExitReason      string  `json:"exit_reason"`
	WorkflowName    string  `json:"workflow_name"`
	WorkflowVersion string  `json:"workflow_version"`
	ParameterSetID  string  `json:"parameter_set_id"`
}

type ListTradesResponse struct {
	Partition string      `json:"partition"`
	RunID     string      `json:"run_id,omitempty"`
	Limit     int         `json:"limit"`
	Offset    int         `json:"offset"`
	Items     []TradeItem `json:"items"`
}

type EquityPoint struct {
	Time     string  `json:"time"`
	TradeID  string  `json:"trade_id"`
	Equity   float64 `json:"equity"`
	Drawdown float64 `json:"drawdown"`
}

type EquitySeriesResponse struct {
	Partition   string        `json:"partition"`
	RunID       string        `json:"run_id,omitempty"`
	MaxDrawdown float64       `json:"max_drawdown"`
	TotalNetPnL float64       `json:"total_net_pnl"`
	Points      []EquityPoint `json:"points"`
}
