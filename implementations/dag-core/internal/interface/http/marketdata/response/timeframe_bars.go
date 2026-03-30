package response

type BackfillTimeframeBarsResponse struct {
	Status        string `json:"status"`
	Entrypoint    string `json:"entrypoint"`
	SymbolID      int64  `json:"symbol_id"`
	TimeframeCode string `json:"timeframe_code"`
	From          string `json:"from"`
	To            string `json:"to"`
	ChunkCount    int    `json:"chunk_count"`
	BarCount      int    `json:"bar_count"`
}

type BackfillTimeframeBarsWorkflowResponse struct {
	Status     string                 `json:"status"`
	Entrypoint string                 `json:"entrypoint"`
	Symbol     string                 `json:"symbol,omitempty"`
	RunID      string                 `json:"run_id"`
	Marketdata *BackfillWorkflowInput `json:"marketdata,omitempty"`
}

type BackfillWorkflowInput struct {
	SymbolID      int64  `json:"symbol_id,omitempty"`
	SymbolCode    string `json:"symbol_code,omitempty"`
	TimeframeCode string `json:"timeframe_code,omitempty"`
	From          string `json:"from,omitempty"`
	To            string `json:"to,omitempty"`
}
