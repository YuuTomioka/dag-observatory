package response

type BacktestMarketdataInput struct {
	SymbolCode    string `json:"symbol_code"`
	TimeframeCode string `json:"timeframe_code"`
	From          string `json:"from"`
	To            string `json:"to"`
	Source        string `json:"source"`
	Timezone      string `json:"timezone"`
	GapHandling   string `json:"gap_handling"`
}

type BacktestRunResponse struct {
	Status     string                  `json:"status"`
	Entrypoint string                  `json:"entrypoint"`
	RunID      string                  `json:"run_id"`
	Partition  string                  `json:"partition"`
	Symbol     string                  `json:"symbol,omitempty"`
	Mode       string                  `json:"mode"`
	Marketdata BacktestMarketdataInput `json:"marketdata"`
}
