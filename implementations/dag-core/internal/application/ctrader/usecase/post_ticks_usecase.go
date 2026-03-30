package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	marketdatarepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

type TickInput struct {
	Time marketdata.UTCTime
	Bid  int64
	Ask  int64
}

type PostTicksRequest struct {
	Symbol     string
	PriceScale int
	Ticks      []TickInput

	RequestID string
	Day       string
	BatchSeq  int
}

type PostTicksResult struct {
	Symbol   string
	SymbolID int64
	Upserted int
}

type PostTicksUsecase struct {
	UnitOfWork marketdatarepository.UnitOfWork
}

func (u *PostTicksUsecase) Execute(ctx context.Context, req PostTicksRequest) (PostTicksResult, error) {
	if u == nil || u.UnitOfWork == nil {
		return PostTicksResult{}, marketdatarepository.ErrNotConfigured
	}

	req.Symbol = strings.TrimSpace(strings.ToUpper(req.Symbol))
	if req.Symbol == "" {
		return PostTicksResult{}, fmt.Errorf("%w: symbol is required", marketdatarepository.ErrInvalidArgument)
	}
	if req.PriceScale < 0 || req.PriceScale > 18 {
		return PostTicksResult{}, fmt.Errorf("%w: price_scale must be between 0 and 18", marketdatarepository.ErrInvalidArgument)
	}
	var dayUTC time.Time
	if strings.TrimSpace(req.Day) != "" {
		parsedDay, err := time.Parse("2006-01-02", req.Day)
		if err != nil {
			return PostTicksResult{}, fmt.Errorf("%w: day must be yyyy-mm-dd", marketdatarepository.ErrInvalidArgument)
		}
		dayUTC = parsedDay.UTC()
	}
	if len(req.Ticks) == 0 {
		return PostTicksResult{}, fmt.Errorf("%w: ticks must not be empty", marketdatarepository.ErrInvalidArgument)
	}
	if req.BatchSeq < 0 {
		return PostTicksResult{}, fmt.Errorf("%w: batch_seq must be >= 0", marketdatarepository.ErrInvalidArgument)
	}

	var result PostTicksResult
	err := u.UnitOfWork.Do(ctx, func(repos marketdatarepository.Repositories) error {
		symbol, err := repos.Symbols().GetByCode(ctx, req.Symbol)
		if err != nil {
			return err
		}
		if int(symbol.PriceScale) != req.PriceScale {
			return fmt.Errorf(
				"%w: price_scale mismatch for symbol=%s expected=%d actual=%d",
				marketdatarepository.ErrConflict,
				req.Symbol,
				symbol.PriceScale,
				req.PriceScale,
			)
		}

		ticks := make([]marketdata.Tick, 0, len(req.Ticks))
		for i, item := range req.Ticks {
			if item.Time.IsZero() {
				return fmt.Errorf("%w: ticks[%d].time is required", marketdatarepository.ErrInvalidArgument, i)
			}
			if !dayUTC.IsZero() {
				ts := item.Time.Time().UTC()
				if ts.Year() != dayUTC.Year() || ts.Month() != dayUTC.Month() || ts.Day() != dayUTC.Day() {
					return fmt.Errorf(
						"%w: ticks[%d].time=%s is outside day=%s",
						marketdatarepository.ErrInvalidArgument,
						i,
						item.Time.String(),
						req.Day,
					)
				}
			}
			ticks = append(ticks, marketdata.Tick{
				SymbolID: symbol.ID,
				Time:     item.Time,
				Bid:      marketdata.NewPriceFromRaw(item.Bid),
				Ask:      marketdata.NewPriceFromRaw(item.Ask),
			})
		}

		if err := repos.Ticks().BulkUpsert(ctx, ticks); err != nil {
			return err
		}

		result = PostTicksResult{
			Symbol:   symbol.Code,
			SymbolID: int64(symbol.ID),
			Upserted: len(ticks),
		}
		return nil
	})
	if err != nil {
		return PostTicksResult{}, err
	}
	return result, nil
}
