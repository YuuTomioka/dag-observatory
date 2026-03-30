package response

type BackfillTimeframeBarsResponse struct {
	SymbolID      int64  `json:"symbol_id"`
	TimeframeCode string `json:"timeframe_code"`
	From          string `json:"from"`
	To            string `json:"to"`
	ChunkCount    int    `json:"chunk_count"`
	BarCount      int    `json:"bar_count"`
}
