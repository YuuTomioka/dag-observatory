package repository

import (
	"context"

	"dag-observatory/dag-core/internal/domain/marketdata"
)

type TickRepository interface {
	Insert(ctx context.Context, tick marketdata.Tick) error
	Upsert(ctx context.Context, tick marketdata.Tick) error
	BulkUpsert(ctx context.Context, ticks []marketdata.Tick) error
	GetLatestBySymbol(ctx context.Context, symbolID marketdata.SymbolID) (marketdata.Tick, error)
	ListBySymbolAndRange(
		ctx context.Context,
		symbolID marketdata.SymbolID,
		from marketdata.UTCTime,
		to marketdata.UTCTime,
	) ([]marketdata.Tick, error)
	ListBySymbolsAndRange(
		ctx context.Context,
		symbolIDs []marketdata.SymbolID,
		from marketdata.UTCTime,
		to marketdata.UTCTime,
	) ([]marketdata.Tick, error)
	ListByRange(ctx context.Context, from marketdata.UTCTime, to marketdata.UTCTime) ([]marketdata.Tick, error)
}
