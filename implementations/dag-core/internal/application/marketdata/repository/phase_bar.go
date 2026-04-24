package repository

import (
	"context"

	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/domain/marketphase"
)

type PhaseBarRepository interface {
	BulkUpsert(ctx context.Context, bars []marketphase.PhaseBar) error
	DeleteBySymbolPhaseAndRange(
		ctx context.Context,
		symbolID marketdata.SymbolID,
		phaseCode marketphase.PhaseCode,
		from marketdata.UTCTime,
		to marketdata.UTCTime,
	) error
	GetLatestBySymbolAndPhase(
		ctx context.Context,
		symbolID marketdata.SymbolID,
		phaseCode marketphase.PhaseCode,
	) (marketphase.PhaseBar, error)
	ListBySymbolPhaseAndRange(
		ctx context.Context,
		symbolID marketdata.SymbolID,
		phaseCode marketphase.PhaseCode,
		from marketdata.UTCTime,
		to marketdata.UTCTime,
	) ([]marketphase.PhaseBar, error)
}
