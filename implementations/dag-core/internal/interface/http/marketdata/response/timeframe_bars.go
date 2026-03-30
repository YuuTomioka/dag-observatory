package response

type BackfillTimeframeBarsResponse struct {
	Status        string `json:"status" example:"completed"`
	Entrypoint    string `json:"entrypoint" example:"direct"`
	SymbolID      int64  `json:"symbol_id" example:"7"`
	TimeframeCode string `json:"timeframe_code" example:"M1"`
	From          string `json:"from" example:"2026-03-01T00:00:00Z"`
	To            string `json:"to" example:"2026-03-01T02:00:00Z"`
	ChunkCount    int    `json:"chunk_count" example:"2"`
	BarCount      int    `json:"bar_count" example:"120"`
}

type BackfillTimeframeBarsWorkflowResponse struct {
	Status     string                 `json:"status" example:"enqueued"`
	Entrypoint string                 `json:"entrypoint" example:"workflow"`
	Symbol     string                 `json:"symbol,omitempty" example:"USDJPY"`
	RunID      string                 `json:"run_id" example:"7c0f8eb6-99f1-4d79-bf4c-e1a7ca3e5f76"`
	Marketdata *BackfillWorkflowInput `json:"marketdata,omitempty"`
}

type BackfillWorkflowInput struct {
	SymbolID      int64  `json:"symbol_id,omitempty" example:"7"`
	SymbolCode    string `json:"symbol_code,omitempty" example:"USDJPY"`
	TimeframeCode string `json:"timeframe_code,omitempty" example:"M1"`
	From          string `json:"from,omitempty" example:"2026-03-01T00:00:00Z"`
	To            string `json:"to,omitempty" example:"2026-03-01T02:00:00Z"`
}
