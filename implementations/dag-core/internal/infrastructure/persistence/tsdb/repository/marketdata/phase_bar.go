package marketdata

import (
	"context"

	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	domainmarketdata "dag-observatory/dag-core/internal/domain/marketdata"
	domainmarketphase "dag-observatory/dag-core/internal/domain/marketphase"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	mapper "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/mapper/marketdata"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
)

type PhaseBarRepository struct {
	queries query.Querier
}

func NewPhaseBarRepository(client *tsdb.Client) *PhaseBarRepository {
	var q query.Querier
	if client != nil && client.Pool != nil {
		q = query.New(client.Pool)
	}
	return NewPhaseBarRepositoryWithQuerier(q)
}

func NewPhaseBarRepositoryWithQuerier(queries query.Querier) *PhaseBarRepository {
	return &PhaseBarRepository{queries: queries}
}

func (r *PhaseBarRepository) BulkUpsert(ctx context.Context, bars []domainmarketphase.PhaseBar) error {
	if err := r.validate(); err != nil {
		return err
	}
	if len(bars) == 0 {
		return nil
	}
	params, err := mapper.ToBulkUpsertPhaseBarsParams(bars)
	if err != nil {
		return err
	}
	return mapRepositoryError(r.queries.BulkUpsertPhaseBars(ctx, params))
}

func (r *PhaseBarRepository) DeleteBySymbolPhaseAndRange(
	ctx context.Context,
	symbolID domainmarketdata.SymbolID,
	phaseID domainmarketphase.PhaseID,
	from domainmarketdata.UTCTime,
	to domainmarketdata.UTCTime,
) error {
	if err := r.validate(); err != nil {
		return err
	}
	params, err := mapper.ToDeletePhaseBarsBySymbolPhaseAndRangeParams(symbolID, phaseID, from, to)
	if err != nil {
		return err
	}
	return mapRepositoryError(r.queries.DeletePhaseBarsBySymbolPhaseAndRange(ctx, params))
}

func (r *PhaseBarRepository) GetLatestBySymbolAndPhase(
	ctx context.Context,
	symbolID domainmarketdata.SymbolID,
	phaseID domainmarketphase.PhaseID,
) (domainmarketphase.PhaseBar, error) {
	if err := r.validate(); err != nil {
		return domainmarketphase.PhaseBar{}, err
	}
	row, err := r.queries.GetLatestPhaseBarBySymbolAndPhase(ctx, mapper.ToGetLatestPhaseBarBySymbolAndPhaseParams(symbolID, phaseID))
	if err != nil {
		return domainmarketphase.PhaseBar{}, mapRepositoryError(err)
	}
	return mapper.PhaseBarFromRow(row)
}

func (r *PhaseBarRepository) ListBySymbolPhaseAndRange(
	ctx context.Context,
	symbolID domainmarketdata.SymbolID,
	phaseID domainmarketphase.PhaseID,
	from domainmarketdata.UTCTime,
	to domainmarketdata.UTCTime,
) ([]domainmarketphase.PhaseBar, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	params, err := mapper.ToListPhaseBarsBySymbolPhaseAndRangeParams(symbolID, phaseID, from, to)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListPhaseBarsBySymbolPhaseAndRange(ctx, params)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return mapper.PhaseBarsFromRows(rows)
}

func (r *PhaseBarRepository) validate() error {
	if r == nil || r.queries == nil {
		return apprepository.ErrNotConfigured
	}
	return nil
}

var _ apprepository.PhaseBarRepository = (*PhaseBarRepository)(nil)
