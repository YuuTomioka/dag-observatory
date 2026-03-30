package request

import "dag-observatory/dag-core/internal/domain/marketdata"

type BackfillTimeframeBarsRequest struct {
	SymbolID      int64              `json:"symbol_id,omitempty" example:"7"`
	SymbolCode    string             `json:"symbol_code,omitempty" example:"USDJPY"`
	TimeframeCode string             `json:"timeframe_code" example:"M1"`
	From          marketdata.UTCTime `json:"from" swaggertype:"string" example:"2026-03-01T00:00:00Z"`
	To            marketdata.UTCTime `json:"to" swaggertype:"string" example:"2026-03-01T02:00:00Z"`
	Mode          string             `json:"mode,omitempty" example:"replace_range"`
	ChunkSizeBars int                `json:"chunk_size_bars,omitempty" example:"1000"`
}
