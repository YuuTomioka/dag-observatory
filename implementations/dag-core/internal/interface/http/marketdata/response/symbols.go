package response

type SymbolItem struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	Base        string `json:"base"`
	Quote       string `json:"quote"`
	PriceScale  uint8  `json:"price_scale"`
	TickSizeRaw int64  `json:"tick_size_raw"`
	PipSizeRaw  int64  `json:"pip_size_raw"`
}

type ListSymbolsResponse struct {
	Items []SymbolItem `json:"items"`
}
