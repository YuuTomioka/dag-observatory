package marketdata

import (
	"context"
	"fmt"

	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
)

type Symbol struct {
	ID          int64
	Code        string
	Base        string
	Quote       string
	PriceScale  int16
	TickSizeRaw int64
	PipSizeRaw  int64
}

type CreateSymbolInput struct {
	Code        string
	Base        string
	Quote       string
	PriceScale  int16
	TickSizeRaw int64
	PipSizeRaw  int64
}

type SymbolRepository struct {
	client  *tsdb.Client
	queries *query.Queries
}

func NewSymbolRepository(client *tsdb.Client) *SymbolRepository {
	var q *query.Queries
	if client != nil && client.Pool != nil {
		q = query.New(client.Pool)
	}
	return &SymbolRepository{
		client:  client,
		queries: q,
	}
}

func (r *SymbolRepository) Create(ctx context.Context, input CreateSymbolInput) (Symbol, error) {
	if err := r.validate(); err != nil {
		return Symbol{}, err
	}
	row, err := r.queries.CreateSymbol(ctx, query.CreateSymbolParams{
		Code:        input.Code,
		Base:        input.Base,
		Quote:       input.Quote,
		PriceScale:  input.PriceScale,
		TickSizeRaw: input.TickSizeRaw,
		PipSizeRaw:  input.PipSizeRaw,
	})
	if err != nil {
		return Symbol{}, err
	}
	return mapSymbol(row), nil
}

func (r *SymbolRepository) GetByID(ctx context.Context, id int64) (Symbol, error) {
	if err := r.validate(); err != nil {
		return Symbol{}, err
	}
	row, err := r.queries.GetSymbolByID(ctx, id)
	if err != nil {
		return Symbol{}, err
	}
	return mapSymbol(row), nil
}

func (r *SymbolRepository) GetByCode(ctx context.Context, code string) (Symbol, error) {
	if err := r.validate(); err != nil {
		return Symbol{}, err
	}
	row, err := r.queries.GetSymbolByCode(ctx, code)
	if err != nil {
		return Symbol{}, err
	}
	return mapSymbol(row), nil
}

func (r *SymbolRepository) List(ctx context.Context) ([]Symbol, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListSymbols(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]Symbol, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSymbol(row))
	}
	return out, nil
}

func (r *SymbolRepository) validate() error {
	if r == nil || r.client == nil || r.client.Pool == nil || r.queries == nil {
		return fmt.Errorf("tsdb: client not configured")
	}
	return nil
}

func mapSymbol(row query.Symbol) Symbol {
	return Symbol{
		ID:          row.ID,
		Code:        row.Code,
		Base:        row.Base,
		Quote:       row.Quote,
		PriceScale:  row.PriceScale,
		TickSizeRaw: row.TickSizeRaw,
		PipSizeRaw:  row.PipSizeRaw,
	}
}
