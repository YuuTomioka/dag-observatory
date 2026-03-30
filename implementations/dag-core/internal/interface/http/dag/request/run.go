package request

import "dag-observatory/dag-core/internal/domain/marketdata"

type MarketdataInput struct {
	SymbolID      int64              `json:"symbol_id,omitempty" example:"7"`
	SymbolCode    string             `json:"symbol_code,omitempty" example:"USDJPY"`
	TimeframeCode string             `json:"timeframe_code,omitempty" example:"M1"`
	From          marketdata.UTCTime `json:"from,omitempty" swaggertype:"string" example:"2026-03-01T00:00:00Z"`
	To            marketdata.UTCTime `json:"to,omitempty" swaggertype:"string" example:"2026-03-01T02:00:00Z"`
}

type RunRequest struct {
	Symbol     string           `json:"symbol" example:"USDJPY"`
	Mode       string           `json:"mode" example:"normal"`
	Bars       []float64        `json:"bars,omitempty"`
	Marketdata *MarketdataInput `json:"marketdata,omitempty"`
}
