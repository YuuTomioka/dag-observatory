package repository

import (
	"context"

	"dag-observatory/dag-core/internal/domain/marketdata"
)

type CreateSymbolInput struct {
	Code        string
	Base        string
	Quote       string
	PriceScale  uint8
	TickSizeRaw int64
	PipSizeRaw  int64
}

type SymbolRepository interface {
	Create(ctx context.Context, input CreateSymbolInput) (marketdata.Symbol, error)
	GetByID(ctx context.Context, id marketdata.SymbolID) (marketdata.Symbol, error)
	GetByCode(ctx context.Context, code string) (marketdata.Symbol, error)
	List(ctx context.Context) ([]marketdata.Symbol, error)
}
