package repository

import (
	"context"

	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/domain/marketdata/ohlc/timeframe"
)

type TimeframeBarRepository interface {
	BulkUpsert(ctx context.Context, bars []timeframe.TimeframeBar) error
	DeleteBySymbolTimeframeAndRange(
		ctx context.Context,
		symbolID marketdata.SymbolID,
		timeframeCode timeframe.TimeframeCode,
		from marketdata.UTCTime,
		to marketdata.UTCTime,
	) error
	GetLatestBySymbolAndTimeframe(
		ctx context.Context,
		symbolID marketdata.SymbolID,
		timeframeCode timeframe.TimeframeCode,
	) (timeframe.TimeframeBar, error)
	ListBySymbolTimeframeAndRange(
		ctx context.Context,
		symbolID marketdata.SymbolID,
		timeframeCode timeframe.TimeframeCode,
		from marketdata.UTCTime,
		to marketdata.UTCTime,
	) ([]timeframe.TimeframeBar, error)
}
