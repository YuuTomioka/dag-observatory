package request

type UpsertTicksBulkRequest struct {
	Ticks []TickItem `json:"ticks"`
}

type TickItem struct {
	SymbolID int64  `json:"symbol_id"`
	Time     string `json:"time"`
	BidRaw   int64  `json:"bid_raw"`
	AskRaw   int64  `json:"ask_raw"`
}
