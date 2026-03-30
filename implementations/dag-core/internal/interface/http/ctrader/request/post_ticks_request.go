package request

type TickRequest struct {
	Time string `json:"time" validate:"required"`
	Bid  int    `json:"bid" validate:"required"`
	Ask  int    `json:"ask" validate:"required"`
}

type PostTicksRequest struct {
	Symbol     string        `json:"symbol" validate:"required"`
	PriceScale int           `json:"price_scale" validate:"required"`
	Ticks      []TickRequest `json:"ticks" validate:"required,dive"`
	RequestID  string        `json:"request_id"`
	Day        string        `json:"day"`
	BatchSeq   int           `json:"batch_seq"`
}
