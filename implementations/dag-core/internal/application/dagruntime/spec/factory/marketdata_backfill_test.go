package factory

import (
	"context"
	"testing"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	marketdatarepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/marketdata"
	artifactinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
)

type fakeFactoryRepositories struct {
	symbols       marketdatarepository.SymbolRepository
	ticks         marketdatarepository.TickRepository
	timeframeBars marketdatarepository.TimeframeBarRepository
}

func (r fakeFactoryRepositories) Symbols() marketdatarepository.SymbolRepository {
	return r.symbols
}

func (r fakeFactoryRepositories) Ticks() marketdatarepository.TickRepository {
	return r.ticks
}

func (r fakeFactoryRepositories) TimeframeBars() marketdatarepository.TimeframeBarRepository {
	return r.timeframeBars
}

type fakeFactoryUnitOfWork struct {
	repos marketdatarepository.Repositories
}

func (u fakeFactoryUnitOfWork) Do(ctx context.Context, fn func(repos marketdatarepository.Repositories) error) error {
	return fn(u.repos)
}

func (u fakeFactoryUnitOfWork) DoReadOnly(ctx context.Context, fn func(repos marketdatarepository.Repositories) error) error {
	return fn(u.repos)
}

type fakeFactorySymbolRepo struct {
	symbol marketdata.Symbol
}

func (r fakeFactorySymbolRepo) Create(ctx context.Context, input marketdatarepository.CreateSymbolInput) (marketdata.Symbol, error) {
	return marketdata.Symbol{}, nil
}
func (r fakeFactorySymbolRepo) GetByID(ctx context.Context, id marketdata.SymbolID) (marketdata.Symbol, error) {
	return marketdata.Symbol{}, nil
}
func (r fakeFactorySymbolRepo) GetByCode(ctx context.Context, code string) (marketdata.Symbol, error) {
	return r.symbol, nil
}
func (r fakeFactorySymbolRepo) List(ctx context.Context) ([]marketdata.Symbol, error) {
	return nil, nil
}

type fakeFactoryTickRepo struct {
	ticks []marketdata.Tick
}

func (r fakeFactoryTickRepo) Insert(ctx context.Context, tick marketdata.Tick) error { return nil }
func (r fakeFactoryTickRepo) Upsert(ctx context.Context, tick marketdata.Tick) error { return nil }
func (r fakeFactoryTickRepo) BulkUpsert(ctx context.Context, ticks []marketdata.Tick) error {
	return nil
}
func (r fakeFactoryTickRepo) GetLatestBySymbol(ctx context.Context, symbolID marketdata.SymbolID) (marketdata.Tick, error) {
	return marketdata.Tick{}, nil
}
func (r fakeFactoryTickRepo) ListBySymbolAndRange(
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
func (r fakeFactoryTickRepo) ListBySymbolsAndRange(
	ctx context.Context,
	symbolIDs []marketdata.SymbolID,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]marketdata.Tick, error) {
	return nil, nil
}
func (r fakeFactoryTickRepo) ListByRange(ctx context.Context, from marketdata.UTCTime, to marketdata.UTCTime) ([]marketdata.Tick, error) {
	return nil, nil
}

type fakeFactoryTimeframeBarRepo struct {
	deletes []backfillChunk
	bars    []marketdata.TimeframeBar
}

func (r *fakeFactoryTimeframeBarRepo) BulkUpsert(ctx context.Context, bars []marketdata.TimeframeBar) error {
	r.bars = append(r.bars, bars...)
	return nil
}
func (r *fakeFactoryTimeframeBarRepo) DeleteBySymbolTimeframeAndRange(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode marketdata.TimeframeCode,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) error {
	r.deletes = append(r.deletes, backfillChunk{From: from, To: to})
	return nil
}
func (r *fakeFactoryTimeframeBarRepo) GetLatestBySymbolAndTimeframe(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode marketdata.TimeframeCode,
) (marketdata.TimeframeBar, error) {
	return marketdata.TimeframeBar{}, nil
}
func (r *fakeFactoryTimeframeBarRepo) ListBySymbolTimeframeAndRange(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode marketdata.TimeframeCode,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]marketdata.TimeframeBar, error) {
	return nil, nil
}

func TestNewBuiltinRegistryWithDependenciesRegistersMarketdataKinds(t *testing.T) {
	t.Parallel()

	registry, err := NewBuiltinRegistryWithDependencies(Dependencies{
		MarketDataUnitOfWork: fakeFactoryUnitOfWork{},
	})
	if err != nil {
		t.Fatalf("new builtin registry with deps: %v", err)
	}

	for _, kind := range []string{
		"marketdata_timeframe_bar_backfill",
		"marketdata_resolve_symbol",
		"marketdata_plan_windows",
		"marketdata_load_ticks_for_chunk",
		"marketdata_aggregate_timeframe_bars",
		"marketdata_persist_timeframe_bars",
	} {
		if !registry.Has(kind) {
			t.Fatalf("expected kind %q to be registered", kind)
		}
	}
}

func TestMarketdataBackfillFiveNodeRun(t *testing.T) {
	t.Parallel()

	barRepo := &fakeFactoryTimeframeBarRepo{}
	uow := fakeFactoryUnitOfWork{
		repos: fakeFactoryRepositories{
			symbols: fakeFactorySymbolRepo{
				symbol: marketdata.Symbol{ID: 7, Code: "USDJPY"},
			},
			ticks: fakeFactoryTickRepo{
				ticks: []marketdata.Tick{
					{SymbolID: 7, Time: marketdata.MustParseUTCTime("2026-03-01T00:00:10Z"), Bid: 1000, Ask: 1002},
					{SymbolID: 7, Time: marketdata.MustParseUTCTime("2026-03-01T00:00:20Z"), Bid: 1004, Ask: 1006},
					{SymbolID: 7, Time: marketdata.MustParseUTCTime("2026-03-01T00:01:05Z"), Bid: 1010, Ask: 1014},
				},
			},
			timeframeBars: barRepo,
		},
	}

	registry, err := NewBuiltinRegistryWithDependencies(Dependencies{MarketDataUnitOfWork: uow})
	if err != nil {
		t.Fatalf("new builtin registry with deps: %v", err)
	}

	nodes := []spec.NodeSpec{
		{ID: "resolve", Kind: "marketdata_resolve_symbol"},
		{ID: "plan", Kind: "marketdata_plan_windows", Config: map[string]any{"resolve_node_id": "resolve", "chunk_size_bars": 1}},
		{ID: "load", Kind: "marketdata_load_ticks_for_chunk", Config: map[string]any{"plan_node_id": "plan"}},
		{ID: "aggregate", Kind: "marketdata_aggregate_timeframe_bars", Config: map[string]any{"load_node_id": "load"}},
		{ID: "persist", Kind: "marketdata_persist_timeframe_bars", Config: map[string]any{"plan_node_id": "plan", "aggregate_node_id": "aggregate", "mode": "replace_range"}},
	}

	artifacts := artifactinfra.NewMemoryStore()
	writer := artifacts
	artifact.Set(writer, dagruntimeusecase.InputKeyMarketdataSymbolCode, "USDJPY")
	artifact.Set(writer, dagruntimeusecase.InputKeyMarketdataTimeframeCode, "M1")
	artifact.Set(writer, dagruntimeusecase.InputKeyMarketdataFrom, marketdata.MustParseUTCTime("2026-03-01T00:00:00Z"))
	artifact.Set(writer, dagruntimeusecase.InputKeyMarketdataTo, marketdata.MustParseUTCTime("2026-03-01T00:02:00Z"))

	txn := stateinfra.NewMemoryStore().BeginTxn("test")
	view := artifacts.View()

	for _, nodeSpec := range nodes {
		built, err := registry.Build(nodeSpec)
		if err != nil {
			t.Fatalf("build node %s: %v", nodeSpec.ID, err)
		}
		if err := built.Run(context.Background(), view, writer, txn); err != nil {
			t.Fatalf("run node %s: %v", nodeSpec.ID, err)
		}
		view = artifacts.View()
	}

	result := artifact.MustGet(view, timeframeBarBackfillResultKey("persist"))
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
		t.Fatalf("expected 2 persisted bars, got %d", len(barRepo.bars))
	}
	if got := barRepo.bars[0].Open.Raw(); got != 1001 {
		t.Fatalf("expected first bar open=1001, got %d", got)
	}
	if got := barRepo.bars[0].Close.Raw(); got != 1005 {
		t.Fatalf("expected first bar close=1005, got %d", got)
	}
}
