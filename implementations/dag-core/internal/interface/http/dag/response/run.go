package response

type MarketdataInput struct {
	SymbolID      int64  `json:"symbol_id,omitempty" example:"7"`
	SymbolCode    string `json:"symbol_code,omitempty" example:"USDJPY"`
	TimeframeCode string `json:"timeframe_code,omitempty" example:"M1"`
	From          string `json:"from,omitempty" example:"2026-03-01T00:00:00Z"`
	To            string `json:"to,omitempty" example:"2026-03-01T02:00:00Z"`
}

type RunResponse struct {
	Status     string           `json:"status" example:"enqueued"`
	Entrypoint string           `json:"entrypoint" example:"workflow"`
	Symbol     string           `json:"symbol,omitempty" example:"USDJPY"`
	RunID      string           `json:"run_id" example:"7c0f8eb6-99f1-4d79-bf4c-e1a7ca3e5f76"`
	Marketdata *MarketdataInput `json:"marketdata,omitempty"`
}
