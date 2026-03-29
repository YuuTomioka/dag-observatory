package marketdata

import (
	"context"

	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	domainmarketdata "dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	mapper "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/mapper/marketdata"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
)

type TickRepository struct {
	queries query.Querier
}

func NewTickRepository(client *tsdb.Client) *TickRepository {
	var q query.Querier
	if client != nil && client.Pool != nil {
		q = query.New(client.Pool)
	}
	return NewTickRepositoryWithQuerier(q)
}

func NewTickRepositoryWithQuerier(queries query.Querier) *TickRepository {
	return &TickRepository{queries: queries}
}

func (r *TickRepository) Insert(ctx context.Context, tick domainmarketdata.Tick) error {
	if err := r.validate(); err != nil {
		return err
	}
	params, err := mapper.ToInsertTickParams(tick)
	if err != nil {
		return err
	}
	return mapRepositoryError(r.queries.InsertTick(ctx, params))
}

func (r *TickRepository) Upsert(ctx context.Context, tick domainmarketdata.Tick) error {
	if err := r.validate(); err != nil {
		return err
	}
	params, err := mapper.ToUpsertTickParams(tick)
	if err != nil {
		return err
	}
	return mapRepositoryError(r.queries.UpsertTick(ctx, params))
}

func (r *TickRepository) BulkUpsert(ctx context.Context, ticks []domainmarketdata.Tick) error {
	if err := r.validate(); err != nil {
		return err
	}
	if len(ticks) == 0 {
		return nil
	}
	params, err := mapper.ToBulkUpsertTickParams(ticks)
	if err != nil {
		return err
	}
	return mapRepositoryError(r.queries.BulkUpsertTicks(ctx, params))
}

func (r *TickRepository) GetLatestBySymbol(
	ctx context.Context,
	symbolID domainmarketdata.SymbolID,
) (domainmarketdata.Tick, error) {
	if err := r.validate(); err != nil {
		return domainmarketdata.Tick{}, err
	}
	row, err := r.queries.GetLatestTickBySymbol(ctx, int64(symbolID))
	if err != nil {
		return domainmarketdata.Tick{}, mapRepositoryError(err)
	}
	mapped, err := mapper.TickFromRow(row)
	if err != nil {
		return domainmarketdata.Tick{}, err
	}
	return mapped, nil
}

func (r *TickRepository) ListBySymbolAndRange(
	ctx context.Context,
	symbolID domainmarketdata.SymbolID,
	from domainmarketdata.UTCTime,
	to domainmarketdata.UTCTime,
) ([]domainmarketdata.Tick, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	params, err := mapper.ToListTicksBySymbolAndRangeParams(symbolID, from, to)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListTicksBySymbolAndRange(ctx, params)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return mapper.TicksFromRows(rows)
}

func (r *TickRepository) ListBySymbolsAndRange(
	ctx context.Context,
	symbolIDs []domainmarketdata.SymbolID,
	from domainmarketdata.UTCTime,
	to domainmarketdata.UTCTime,
) ([]domainmarketdata.Tick, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	params, err := mapper.ToListTicksBySymbolsAndRangeParams(symbolIDs, from, to)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListTicksBySymbolsAndRange(ctx, params)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return mapper.TicksFromRows(rows)
}

func (r *TickRepository) ListByRange(
	ctx context.Context,
	from domainmarketdata.UTCTime,
	to domainmarketdata.UTCTime,
) ([]domainmarketdata.Tick, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	params, err := mapper.ToListTicksByRangeParams(from, to)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListTicksByRange(ctx, params)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return mapper.TicksFromRows(rows)
}

func (r *TickRepository) validate() error {
	if r == nil || r.queries == nil {
		return apprepository.ErrNotConfigured
	}
	return nil
}

var _ apprepository.TickRepository = (*TickRepository)(nil)
