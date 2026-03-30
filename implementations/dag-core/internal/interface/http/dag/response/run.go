package response

type MarketdataInput struct {
	SymbolID      int64  `json:"symbol_id,omitempty"`
	SymbolCode    string `json:"symbol_code,omitempty"`
	TimeframeCode string `json:"timeframe_code,omitempty"`
	From          string `json:"from,omitempty"`
	To            string `json:"to,omitempty"`
}

type RunResponse struct {
	Status     string           `json:"status"`
	Entrypoint string           `json:"entrypoint"`
	Symbol     string           `json:"symbol,omitempty"`
	RunID      string           `json:"run_id"`
	Marketdata *MarketdataInput `json:"marketdata,omitempty"`
}
