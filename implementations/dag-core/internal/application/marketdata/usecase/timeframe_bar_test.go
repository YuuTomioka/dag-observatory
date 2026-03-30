package usecase

import (
	"context"
	"testing"

	"dag-observatory/dag-core/internal/application/marketdata/repository"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

type fakeBackfillRepositories struct {
	symbols       repository.SymbolRepository
	ticks         repository.TickRepository
	timeframeBars repository.TimeframeBarRepository
}

func (r fakeBackfillRepositories) Symbols() repository.SymbolRepository {
	return r.symbols
}

func (r fakeBackfillRepositories) Ticks() repository.TickRepository {
	return r.ticks
}

func (r fakeBackfillRepositories) TimeframeBars() repository.TimeframeBarRepository {
	return r.timeframeBars
}

type fakeBackfillUnitOfWork struct {
	repos repository.Repositories
}

func (u fakeBackfillUnitOfWork) Do(ctx context.Context, fn func(repos repository.Repositories) error) error {
	return fn(u.repos)
}

func (u fakeBackfillUnitOfWork) DoReadOnly(ctx context.Context, fn func(repos repository.Repositories) error) error {
	return fn(u.repos)
}

type fakeBackfillSymbolRepo struct {
	symbol marketdata.Symbol
}

func (r fakeBackfillSymbolRepo) Create(ctx context.Context, input repository.CreateSymbolInput) (marketdata.Symbol, error) {
	return marketdata.Symbol{}, nil
}

func (r fakeBackfillSymbolRepo) GetByID(ctx context.Context, id marketdata.SymbolID) (marketdata.Symbol, error) {
	return marketdata.Symbol{}, nil
}

func (r fakeBackfillSymbolRepo) GetByCode(ctx context.Context, code string) (marketdata.Symbol, error) {
	return r.symbol, nil
}

func (r fakeBackfillSymbolRepo) List(ctx context.Context) ([]marketdata.Symbol, error) {
	return nil, nil
}

type fakeBackfillTickRepo struct {
	ticks []marketdata.Tick
}

func (r fakeBackfillTickRepo) Insert(ctx context.Context, tick marketdata.Tick) error { return nil }
func (r fakeBackfillTickRepo) Upsert(ctx context.Context, tick marketdata.Tick) error { return nil }
func (r fakeBackfillTickRepo) BulkUpsert(ctx context.Context, ticks []marketdata.Tick) error {
	return nil
}
func (r fakeBackfillTickRepo) GetLatestBySymbol(ctx context.Context, symbolID marketdata.SymbolID) (marketdata.Tick, error) {
	return marketdata.Tick{}, nil
}
func (r fakeBackfillTickRepo) ListBySymbolAndRange(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]marketdata.Tick, error) {
	out := make([]marketdata.Tick, 0)
	for _, tick := range r.ticks {
		if tick.SymbolID != symbolID {
			continue
		}
		if tick.Time.Compare(from) >= 0 && tick.Time.Compare(to) < 0 {
			out = append(out, tick)
		}
	}
	return out, nil
}
func (r fakeBackfillTickRepo) ListBySymbolsAndRange(
	ctx context.Context,
	symbolIDs []marketdata.SymbolID,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]marketdata.Tick, error) {
	return nil, nil
}
func (r fakeBackfillTickRepo) ListByRange(ctx context.Context, from marketdata.UTCTime, to marketdata.UTCTime) ([]marketdata.Tick, error) {
	return nil, nil
}

type fakeBackfillTimeframeBarRepo struct {
	deletes []deleteCall
	bars    []marketdata.TimeframeBar
}

type deleteCall struct {
	from marketdata.UTCTime
	to   marketdata.UTCTime
}

func (r *fakeBackfillTimeframeBarRepo) BulkUpsert(ctx context.Context, bars []marketdata.TimeframeBar) error {
	r.bars = append(r.bars, bars...)
	return nil
}

func (r *fakeBackfillTimeframeBarRepo) DeleteBySymbolTimeframeAndRange(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode marketdata.TimeframeCode,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) error {
	r.deletes = append(r.deletes, deleteCall{from: from, to: to})
	return nil
}

func (r *fakeBackfillTimeframeBarRepo) GetLatestBySymbolAndTimeframe(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode marketdata.TimeframeCode,
) (marketdata.TimeframeBar, error) {
	return marketdata.TimeframeBar{}, nil
}

func (r *fakeBackfillTimeframeBarRepo) ListBySymbolTimeframeAndRange(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode marketdata.TimeframeCode,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]marketdata.TimeframeBar, error) {
	return nil, nil
}

func TestBackfillTimeframeBarsReplaceRange(t *testing.T) {
	t.Parallel()

	barRepo := &fakeBackfillTimeframeBarRepo{}
	uow := fakeBackfillUnitOfWork{
		repos: fakeBackfillRepositories{
			symbols: fakeBackfillSymbolRepo{
				symbol: marketdata.Symbol{ID: 7, Code: "USDJPY"},
			},
			ticks: fakeBackfillTickRepo{
				ticks: []marketdata.Tick{
					{SymbolID: 7, Time: marketdata.MustParseUTCTime("2026-03-01T00:00:10Z"), Bid: 1000, Ask: 1002},
					{SymbolID: 7, Time: marketdata.MustParseUTCTime("2026-03-01T00:00:20Z"), Bid: 1004, Ask: 1006},
					{SymbolID: 7, Time: marketdata.MustParseUTCTime("2026-03-01T00:01:05Z"), Bid: 1010, Ask: 1014},
				},
			},
			timeframeBars: barRepo,
		},
	}

	uc := BackfillTimeframeBars{UnitOfWork: uow}
	result, err := uc.Execute(context.Background(), BackfillTimeframeBarsRequest{
		SymbolCode:    "usdjpy",
		TimeframeCode: marketdata.TimeframeM1,
		From:          marketdata.MustParseUTCTime("2026-03-01T00:00:00Z"),
		To:            marketdata.MustParseUTCTime("2026-03-01T00:02:00Z"),
		ChunkSizeBars: 1,
	})
	if err != nil {
		t.Fatalf("execute backfill: %v", err)
	}

	if result.SymbolID != 7 {
		t.Fatalf("expected symbol_id=7, got %d", result.SymbolID)
	}
	if result.ChunkCount != 2 {
		t.Fatalf("expected 2 chunks, got %d", result.ChunkCount)
	}
	if result.BarCount != 2 {
		t.Fatalf("expected 2 bars, got %d", result.BarCount)
	}
	if len(barRepo.deletes) != 2 {
		t.Fatalf("expected 2 delete calls, got %d", len(barRepo.deletes))
	}
	if len(barRepo.bars) != 2 {
		t.Fatalf("expected 2 upserted bars, got %d", len(barRepo.bars))
	}
	if got := barRepo.bars[0].Open.Raw(); got != 1001 {
		t.Fatalf("expected first bar open=1001, got %d", got)
	}
	if got := barRepo.bars[0].Close.Raw(); got != 1005 {
		t.Fatalf("expected first bar close=1005, got %d", got)
	}
	if got := int64(barRepo.bars[0].Volume); got != 2 {
		t.Fatalf("expected first bar volume=2, got %d", got)
	}
}
