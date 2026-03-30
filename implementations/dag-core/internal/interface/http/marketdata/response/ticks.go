package response

type UpsertTicksBulkResponse struct {
	Upserted int `json:"upserted"`
}

type TickItem struct {
	SymbolID int64  `json:"symbol_id"`
	Time     string `json:"time"`
	BidRaw   int64  `json:"bid_raw"`
	AskRaw   int64  `json:"ask_raw"`
}

type ListTicksResponse struct {
	Items []TickItem `json:"items"`
}
