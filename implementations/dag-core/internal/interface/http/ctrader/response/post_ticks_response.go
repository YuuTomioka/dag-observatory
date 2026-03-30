package response

type PostTicksResponse struct {
	Status    string `json:"status"`
	Symbol    string `json:"symbol"`
	SymbolID  int64  `json:"symbol_id"`
	Upserted  int    `json:"upserted"`
	RequestID string `json:"request_id,omitempty"`
	BatchSeq  int    `json:"batch_seq"`
}
