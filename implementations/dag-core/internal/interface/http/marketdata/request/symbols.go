package request

type CreateSymbolRequest struct {
	Code        string `json:"code"`
	Base        string `json:"base"`
	Quote       string `json:"quote"`
	PriceScale  uint8  `json:"price_scale"`
	TickSizeRaw int64  `json:"tick_size_raw"`
	PipSizeRaw  int64  `json:"pip_size_raw"`
}
