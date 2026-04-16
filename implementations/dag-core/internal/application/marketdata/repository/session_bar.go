package repository

import (
	"context"

	"dag-observatory/dag-core/internal/domain/marketdata"
)

type SessionBarRepository interface {
	BulkUpsert(ctx context.Context, bars []marketdata.SessionBar) error
	DeleteBySymbolSessionAndDateRange(
		ctx context.Context,
		symbolID marketdata.SymbolID,
		sessionCode marketdata.SessionCode,
		from marketdata.SessionDate,
		to marketdata.SessionDate,
	) error
	GetLatestBySymbolAndSession(
		ctx context.Context,
		symbolID marketdata.SymbolID,
		sessionCode marketdata.SessionCode,
	) (marketdata.SessionBar, error)
	ListBySymbolSessionAndDateRange(
		ctx context.Context,
		symbolID marketdata.SymbolID,
		sessionCode marketdata.SessionCode,
		from marketdata.SessionDate,
		to marketdata.SessionDate,
	) ([]marketdata.SessionBar, error)
}
