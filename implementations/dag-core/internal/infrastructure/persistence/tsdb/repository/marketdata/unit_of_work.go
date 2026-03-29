package marketdata

import (
	"context"
	"fmt"

	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
	"github.com/jackc/pgx/v5"
)

type repositories struct {
	symbols apprepository.SymbolRepository
	ticks   apprepository.TickRepository
}

func (r repositories) Symbols() apprepository.SymbolRepository {
	return r.symbols
}

func (r repositories) Ticks() apprepository.TickRepository {
	return r.ticks
}

type UnitOfWork struct {
	client *tsdb.Client
}

func NewUnitOfWork(client *tsdb.Client) *UnitOfWork {
	return &UnitOfWork{client: client}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(repos apprepository.Repositories) error) error {
	return u.do(ctx, pgx.TxOptions{}, fn)
}

func (u *UnitOfWork) DoReadOnly(ctx context.Context, fn func(repos apprepository.Repositories) error) error {
	return u.do(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly}, fn)
}

func (u *UnitOfWork) do(
	ctx context.Context,
	txOptions pgx.TxOptions,
	fn func(repos apprepository.Repositories) error,
) error {
	if u == nil || u.client == nil || u.client.Pool == nil {
		return apprepository.ErrNotConfigured
	}
	if fn == nil {
		return fmt.Errorf("%w: transaction callback is required", apprepository.ErrInvalidArgument)
	}

	tx, err := u.client.Pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	queries := query.New(tx)
	repos := repositories{
		symbols: NewSymbolRepositoryWithQuerier(queries),
		ticks:   NewTickRepositoryWithQuerier(queries),
	}

	if err := fn(repos); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

var _ apprepository.Repositories = repositories{}
var _ apprepository.UnitOfWork = (*UnitOfWork)(nil)
