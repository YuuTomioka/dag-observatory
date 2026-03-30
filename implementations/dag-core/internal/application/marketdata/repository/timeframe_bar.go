package repository

import (
	"context"

	"dag-observatory/dag-core/internal/domain/marketdata"
)

type TimeframeBarRepository interface {
	BulkUpsert(ctx context.Context, bars []marketdata.TimeframeBar) error
	DeleteBySymbolTimeframeAndRange(
		ctx context.Context,
		symbolID marketdata.SymbolID,
		timeframeCode marketdata.TimeframeCode,
		from marketdata.UTCTime,
		to marketdata.UTCTime,
	) error
	GetLatestBySymbolAndTimeframe(
		ctx context.Context,
		symbolID marketdata.SymbolID,
		timeframeCode marketdata.TimeframeCode,
	) (marketdata.TimeframeBar, error)
	ListBySymbolTimeframeAndRange(
		ctx context.Context,
		symbolID marketdata.SymbolID,
		timeframeCode marketdata.TimeframeCode,
		from marketdata.UTCTime,
		to marketdata.UTCTime,
	) ([]marketdata.TimeframeBar, error)
}
