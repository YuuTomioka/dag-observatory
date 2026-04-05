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
	Items     []TradeItem `json:"items"`
}
