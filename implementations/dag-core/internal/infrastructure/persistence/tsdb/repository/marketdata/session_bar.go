package marketdata

import (
	"context"

	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	domainmarketdata "dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	mapper "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/mapper/marketdata"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
)

type SessionBarRepository struct {
	queries query.Querier
}

func NewSessionBarRepository(client *tsdb.Client) *SessionBarRepository {
	var q query.Querier
	if client != nil && client.Pool != nil {
		q = query.New(client.Pool)
	}
	return NewSessionBarRepositoryWithQuerier(q)
}

func NewSessionBarRepositoryWithQuerier(queries query.Querier) *SessionBarRepository {
	return &SessionBarRepository{queries: queries}
}

func (r *SessionBarRepository) BulkUpsert(ctx context.Context, bars []domainmarketdata.SessionBar) error {
	if err := r.validate(); err != nil {
		return err
	}
	if len(bars) == 0 {
		return nil
	}
	params, err := mapper.ToBulkUpsertSessionBarsParams(bars)
	if err != nil {
		return err
	}
	return mapRepositoryError(r.queries.BulkUpsertSessionBars(ctx, params))
}

func (r *SessionBarRepository) DeleteBySymbolSessionAndDateRange(
	ctx context.Context,
	symbolID domainmarketdata.SymbolID,
	sessionCode domainmarketdata.SessionCode,
	from domainmarketdata.SessionDate,
	to domainmarketdata.SessionDate,
) error {
	if err := r.validate(); err != nil {
		return err
	}
	params, err := mapper.ToDeleteSessionBarsBySymbolSessionAndDateRangeParams(symbolID, sessionCode, from, to)
	if err != nil {
		return err
	}
	return mapRepositoryError(r.queries.DeleteSessionBarsBySymbolSessionAndDateRange(ctx, params))
}

func (r *SessionBarRepository) GetLatestBySymbolAndSession(
	ctx context.Context,
	symbolID domainmarketdata.SymbolID,
	sessionCode domainmarketdata.SessionCode,
) (domainmarketdata.SessionBar, error) {
	if err := r.validate(); err != nil {
		return domainmarketdata.SessionBar{}, err
	}
	row, err := r.queries.GetLatestSessionBarBySymbolAndSession(
		ctx,
		mapper.ToGetLatestSessionBarBySymbolAndSessionParams(symbolID, sessionCode),
	)
	if err != nil {
		return domainmarketdata.SessionBar{}, mapRepositoryError(err)
	}
	return mapper.SessionBarFromRow(row)
}

func (r *SessionBarRepository) ListBySymbolSessionAndDateRange(
	ctx context.Context,
	symbolID domainmarketdata.SymbolID,
	sessionCode domainmarketdata.SessionCode,
	from domainmarketdata.SessionDate,
	to domainmarketdata.SessionDate,
) ([]domainmarketdata.SessionBar, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	params, err := mapper.ToListSessionBarsBySymbolSessionAndDateRangeParams(symbolID, sessionCode, from, to)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListSessionBarsBySymbolSessionAndDateRange(ctx, params)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return mapper.SessionBarsFromRows(rows)
}

func (r *SessionBarRepository) validate() error {
	if r == nil || r.queries == nil {
		return apprepository.ErrNotConfigured
	}
	return nil
}

var _ apprepository.SessionBarRepository = (*SessionBarRepository)(nil)
