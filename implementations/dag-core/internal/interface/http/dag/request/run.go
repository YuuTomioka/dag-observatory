package request

import "dag-observatory/dag-core/internal/domain/marketdata"

type MarketdataInput struct {
	SymbolID      int64              `json:"symbol_id,omitempty"`
	SymbolCode    string             `json:"symbol_code,omitempty"`
	TimeframeCode string             `json:"timeframe_code,omitempty"`
	From          marketdata.UTCTime `json:"from,omitempty"`
	To            marketdata.UTCTime `json:"to,omitempty"`
}

type RunRequest struct {
	Symbol     string           `json:"symbol"`
	Mode       string           `json:"mode"`
	Bars       []float64        `json:"bars,omitempty"`
	Marketdata *MarketdataInput `json:"marketdata,omitempty"`
}
