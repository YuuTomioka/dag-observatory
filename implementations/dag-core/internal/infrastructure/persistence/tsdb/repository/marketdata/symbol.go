package marketdata

import (
	"context"

	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	domainmarketdata "dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	mapper "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/mapper/marketdata"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
)

type SymbolRepository struct {
	queries query.Querier
}

func NewSymbolRepository(client *tsdb.Client) *SymbolRepository {
	var q query.Querier
	if client != nil && client.Pool != nil {
		q = query.New(client.Pool)
	}
	return NewSymbolRepositoryWithQuerier(q)
}

func NewSymbolRepositoryWithQuerier(queries query.Querier) *SymbolRepository {
	return &SymbolRepository{queries: queries}
}

func (r *SymbolRepository) Create(
	ctx context.Context,
	input apprepository.CreateSymbolInput,
) (domainmarketdata.Symbol, error) {
	if err := r.validate(); err != nil {
		return domainmarketdata.Symbol{}, err
	}
	params, err := mapper.ToCreateSymbolParams(input)
	if err != nil {
		return domainmarketdata.Symbol{}, err
	}
	row, err := r.queries.CreateSymbol(ctx, params)
	if err != nil {
		return domainmarketdata.Symbol{}, mapRepositoryError(err)
	}
	mapped, err := mapper.SymbolFromRow(row)
	if err != nil {
		return domainmarketdata.Symbol{}, err
	}
	return mapped, nil
}

func (r *SymbolRepository) GetByID(ctx context.Context, id domainmarketdata.SymbolID) (domainmarketdata.Symbol, error) {
	if err := r.validate(); err != nil {
		return domainmarketdata.Symbol{}, err
	}
	row, err := r.queries.GetSymbolByID(ctx, int64(id))
	if err != nil {
		return domainmarketdata.Symbol{}, mapRepositoryError(err)
	}
	mapped, err := mapper.SymbolFromRow(row)
	if err != nil {
		return domainmarketdata.Symbol{}, err
	}
	return mapped, nil
}

func (r *SymbolRepository) GetByCode(ctx context.Context, code string) (domainmarketdata.Symbol, error) {
	if err := r.validate(); err != nil {
		return domainmarketdata.Symbol{}, err
	}
	row, err := r.queries.GetSymbolByCode(ctx, code)
	if err != nil {
		return domainmarketdata.Symbol{}, mapRepositoryError(err)
	}
	mapped, err := mapper.SymbolFromRow(row)
	if err != nil {
		return domainmarketdata.Symbol{}, err
	}
	return mapped, nil
}

func (r *SymbolRepository) List(ctx context.Context) ([]domainmarketdata.Symbol, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListSymbols(ctx)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return mapper.SymbolsFromRows(rows)
}

func (r *SymbolRepository) validate() error {
	if r == nil || r.queries == nil {
		return apprepository.ErrNotConfigured
	}
	return nil
}

var _ apprepository.SymbolRepository = (*SymbolRepository)(nil)
