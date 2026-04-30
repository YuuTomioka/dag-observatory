package marketdata

import (
	"context"

	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	domainmarketdata "dag-observatory/dag-core/internal/domain/marketdata"
	domaintimeframe "dag-observatory/dag-core/internal/domain/marketdata/ohlc/timeframe"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	mapper "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/mapper/marketdata"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
)

type TimeframeBarRepository struct {
	queries query.Querier
}

func NewTimeframeBarRepository(client *tsdb.Client) *TimeframeBarRepository {
	var q query.Querier
	if client != nil && client.Pool != nil {
		q = query.New(client.Pool)
	}
	return NewTimeframeBarRepositoryWithQuerier(q)
}

func NewTimeframeBarRepositoryWithQuerier(queries query.Querier) *TimeframeBarRepository {
	return &TimeframeBarRepository{queries: queries}
}

func (r *TimeframeBarRepository) BulkUpsert(ctx context.Context, bars []domaintimeframe.TimeframeBar) error {
	if err := r.validate(); err != nil {
		return err
	}
	if len(bars) == 0 {
		return nil
	}
	params, err := mapper.ToBulkUpsertTimeframeBarsParams(bars)
	if err != nil {
		return err
	}
	return mapRepositoryError(r.queries.BulkUpsertTimeframeBars(ctx, params))
}

func (r *TimeframeBarRepository) DeleteBySymbolTimeframeAndRange(
	ctx context.Context,
	symbolID domainmarketdata.SymbolID,
	timeframeCode domaintimeframe.TimeframeCode,
	from domainmarketdata.UTCTime,
	to domainmarketdata.UTCTime,
) error {
	if err := r.validate(); err != nil {
		return err
	}
	params, err := mapper.ToDeleteTimeframeBarsBySymbolTimeframeAndRangeParams(symbolID, timeframeCode, from, to)
	if err != nil {
		return err
	}
	return mapRepositoryError(r.queries.DeleteTimeframeBarsBySymbolTimeframeAndRange(ctx, params))
}

func (r *TimeframeBarRepository) GetLatestBySymbolAndTimeframe(
	ctx context.Context,
	symbolID domainmarketdata.SymbolID,
	timeframeCode domaintimeframe.TimeframeCode,
) (domaintimeframe.TimeframeBar, error) {
	if err := r.validate(); err != nil {
		return domaintimeframe.TimeframeBar{}, err
	}
	row, err := r.queries.GetLatestTimeframeBarBySymbolAndTimeframe(
		ctx,
		mapper.ToGetLatestTimeframeBarBySymbolAndTimeframeParams(symbolID, timeframeCode),
	)
	if err != nil {
		return domaintimeframe.TimeframeBar{}, mapRepositoryError(err)
	}
	return mapper.TimeframeBarFromRow(row)
}

func (r *TimeframeBarRepository) ListBySymbolTimeframeAndRange(
	ctx context.Context,
	symbolID domainmarketdata.SymbolID,
	timeframeCode domaintimeframe.TimeframeCode,
	from domainmarketdata.UTCTime,
	to domainmarketdata.UTCTime,
) ([]domaintimeframe.TimeframeBar, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	params, err := mapper.ToListTimeframeBarsBySymbolTimeframeAndRangeParams(symbolID, timeframeCode, from, to)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListTimeframeBarsBySymbolTimeframeAndRange(ctx, params)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return mapper.TimeframeBarsFromRows(rows)
}

func (r *TimeframeBarRepository) validate() error {
	if r == nil || r.queries == nil {
		return apprepository.ErrNotConfigured
	}
	return nil
}

var _ apprepository.TimeframeBarRepository = (*TimeframeBarRepository)(nil)
