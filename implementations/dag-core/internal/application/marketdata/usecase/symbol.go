package usecase

import (
	"context"
	"fmt"
	"strings"

	"dag-observatory/dag-core/internal/application/marketdata/repository"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

type CreateSymbol struct {
	UnitOfWork repository.UnitOfWork
}

type CreateSymbolRequest struct {
	Code        string
	Base        string
	Quote       string
	PriceScale  uint8
	TickSizeRaw int64
	PipSizeRaw  int64
}

func (u *CreateSymbol) Execute(ctx context.Context, req CreateSymbolRequest) (marketdata.Symbol, error) {
	if u == nil || u.UnitOfWork == nil {
		return marketdata.Symbol{}, repository.ErrNotConfigured
	}
	if err := validateCreateSymbolRequest(req); err != nil {
		return marketdata.Symbol{}, err
	}

	var out marketdata.Symbol
	err := u.UnitOfWork.Do(ctx, func(repos repository.Repositories) error {
		created, err := repos.Symbols().Create(ctx, repository.CreateSymbolInput{
			Code:        strings.ToUpper(req.Code),
			Base:        strings.ToUpper(req.Base),
			Quote:       strings.ToUpper(req.Quote),
			PriceScale:  req.PriceScale,
			TickSizeRaw: req.TickSizeRaw,
			PipSizeRaw:  req.PipSizeRaw,
		})
		if err != nil {
			return err
		}
		out = created
		return nil
	})
	if err != nil {
		return marketdata.Symbol{}, err
	}
	return out, nil
}

type GetSymbolByCode struct {
	UnitOfWork repository.UnitOfWork
}

func (u *GetSymbolByCode) Execute(ctx context.Context, code string) (marketdata.Symbol, error) {
	if u == nil || u.UnitOfWork == nil {
		return marketdata.Symbol{}, repository.ErrNotConfigured
	}
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == "" {
		return marketdata.Symbol{}, fmt.Errorf("%w: code is required", repository.ErrInvalidArgument)
	}

	var out marketdata.Symbol
	err := u.UnitOfWork.DoReadOnly(ctx, func(repos repository.Repositories) error {
		found, err := repos.Symbols().GetByCode(ctx, code)
		if err != nil {
			return err
		}
		out = found
		return nil
	})
	if err != nil {
		return marketdata.Symbol{}, err
	}
	return out, nil
}

type ListSymbols struct {
	UnitOfWork repository.UnitOfWork
}

func (u *ListSymbols) Execute(ctx context.Context) ([]marketdata.Symbol, error) {
	if u == nil || u.UnitOfWork == nil {
		return nil, repository.ErrNotConfigured
	}

	var out []marketdata.Symbol
	err := u.UnitOfWork.DoReadOnly(ctx, func(repos repository.Repositories) error {
		items, err := repos.Symbols().List(ctx)
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

func validateCreateSymbolRequest(req CreateSymbolRequest) error {
	if strings.TrimSpace(req.Code) == "" {
		return fmt.Errorf("%w: code is required", repository.ErrInvalidArgument)
	}
	if strings.TrimSpace(req.Base) == "" {
		return fmt.Errorf("%w: base is required", repository.ErrInvalidArgument)
	}
	if strings.TrimSpace(req.Quote) == "" {
		return fmt.Errorf("%w: quote is required", repository.ErrInvalidArgument)
	}
	if req.TickSizeRaw <= 0 {
		return fmt.Errorf("%w: tick_size_raw must be > 0", repository.ErrInvalidArgument)
	}
	if req.PipSizeRaw <= 0 {
		return fmt.Errorf("%w: pip_size_raw must be > 0", repository.ErrInvalidArgument)
	}
	return nil
}
