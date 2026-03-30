package request

import "dag-observatory/dag-core/internal/domain/marketdata"

type BackfillTimeframeBarsRequest struct {
	SymbolID      int64              `json:"symbol_id,omitempty"`
	SymbolCode    string             `json:"symbol_code,omitempty"`
	TimeframeCode string             `json:"timeframe_code"`
	From          marketdata.UTCTime `json:"from"`
	To            marketdata.UTCTime `json:"to"`
	Mode          string             `json:"mode,omitempty"`
	ChunkSizeBars int                `json:"chunk_size_bars,omitempty"`
}
