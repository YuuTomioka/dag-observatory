package repository

import "context"

type Repositories interface {
	Symbols() SymbolRepository
	Ticks() TickRepository
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(repos Repositories) error) error
	DoReadOnly(ctx context.Context, fn func(repos Repositories) error) error
}
