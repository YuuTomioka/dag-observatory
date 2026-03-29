package usecase

import (
	"context"
	"fmt"

	"dag-observatory/dag-core/internal/application/marketdata/repository"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

type UpsertTicks struct {
	UnitOfWork repository.UnitOfWork
}

func (u *UpsertTicks) Execute(ctx context.Context, ticks []marketdata.Tick) error {
	if u == nil || u.UnitOfWork == nil {
		return repository.ErrNotConfigured
	}
	if len(ticks) == 0 {
		return fmt.Errorf("%w: ticks must not be empty", repository.ErrInvalidArgument)
	}
	for i, tick := range ticks {
		if tick.SymbolID <= 0 {
			return fmt.Errorf("%w: ticks[%d].symbol_id must be > 0", repository.ErrInvalidArgument, i)
		}
		if tick.Time.IsZero() {
			return fmt.Errorf("%w: ticks[%d].time is required", repository.ErrInvalidArgument, i)
		}
	}
	return u.UnitOfWork.Do(ctx, func(repos repository.Repositories) error {
		return repos.Ticks().BulkUpsert(ctx, ticks)
	})
}

type GetLatestTickBySymbol struct {
	UnitOfWork repository.UnitOfWork
}

func (u *GetLatestTickBySymbol) Execute(ctx context.Context, symbolID marketdata.SymbolID) (marketdata.Tick, error) {
	if u == nil || u.UnitOfWork == nil {
		return marketdata.Tick{}, repository.ErrNotConfigured
	}
	if symbolID <= 0 {
		return marketdata.Tick{}, fmt.Errorf("%w: symbol_id must be > 0", repository.ErrInvalidArgument)
	}

	var out marketdata.Tick
	err := u.UnitOfWork.DoReadOnly(ctx, func(repos repository.Repositories) error {
		tick, err := repos.Ticks().GetLatestBySymbol(ctx, symbolID)
		if err != nil {
			return err
		}
		out = tick
		return nil
	})
	if err != nil {
		return marketdata.Tick{}, err
	}
	return out, nil
}

type ListTicksBySymbolAndRange struct {
	UnitOfWork repository.UnitOfWork
}

func (u *ListTicksBySymbolAndRange) Execute(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]marketdata.Tick, error) {
	if u == nil || u.UnitOfWork == nil {
		return nil, repository.ErrNotConfigured
	}
	if symbolID <= 0 {
		return nil, fmt.Errorf("%w: symbol_id must be > 0", repository.ErrInvalidArgument)
	}
	if from.IsZero() || to.IsZero() {
		return nil, fmt.Errorf("%w: from/to are required", repository.ErrInvalidArgument)
	}
	if !from.Before(to) {
		return nil, fmt.Errorf("%w: from must be before to", repository.ErrInvalidArgument)
	}

	var out []marketdata.Tick
	err := u.UnitOfWork.DoReadOnly(ctx, func(repos repository.Repositories) error {
		items, err := repos.Ticks().ListBySymbolAndRange(ctx, symbolID, from, to)
		if err != nil {
			return err
		}
		out = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
