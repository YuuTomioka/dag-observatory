package request

type TickRequest struct {
	Time string `json:"time" validate:"required"`
	Bid  int64  `json:"bid" validate:"required"`
	Ask  int64  `json:"ask" validate:"required"`
}

type PostTicksRequest struct {
	Symbol     string        `json:"symbol" validate:"required"`
	PriceScale int           `json:"price_scale" validate:"required"`
	Ticks      []TickRequest `json:"ticks" validate:"required,dive"`
	RequestID  string        `json:"request_id"`
	Day        string        `json:"day"`
	BatchSeq   int           `json:"batch_seq"`
}
