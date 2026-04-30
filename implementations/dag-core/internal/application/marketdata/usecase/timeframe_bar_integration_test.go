package usecase

import (
	"context"
	"os"
	"testing"

	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/domain/marketdata/ohlc/timeframe"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	marketdatarepo "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/repository/marketdata"
)

func TestBackfillTimeframeBarsPersistsToTSDB(t *testing.T) {
	tsdbURL := os.Getenv("TSDB_URL")
	if tsdbURL == "" {
		t.Skip("TSDB_URL is required for TSDB integration test")
	}

	client, err := tsdb.New(tsdbURL)
	if err != nil {
		t.Fatalf("tsdb.New: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	code := "ITBACKFILLBAR"
	cleanupBackfillIntegrationFixtures(ctx, t, client, code)
	defer cleanupBackfillIntegrationFixtures(ctx, t, client, code)

	uow := marketdatarepo.NewUnitOfWork(client)
	createSymbol := &CreateSymbol{UnitOfWork: uow}
	upsertTicks := &UpsertTicks{UnitOfWork: uow}
	backfill := &BackfillTimeframeBars{UnitOfWork: uow}
	listBars := &ListTimeframeBarsBySymbolTimeframeAndRange{UnitOfWork: uow}

	symbol, err := createSymbol.Execute(ctx, CreateSymbolRequest{
		Code:        code,
		Base:        "ITB",
		Quote:       "JPY",
		PriceScale:  5,
		TickSizeRaw: 1,
		PipSizeRaw:  10,
	})
	if err != nil {
		t.Fatalf("create symbol: %v", err)
	}

	err = upsertTicks.Execute(ctx, []marketdata.Tick{
		{SymbolID: symbol.ID, Time: marketdata.MustParseUTCTime("2026-03-01T00:00:10Z"), Bid: 1000, Ask: 1002},
		{SymbolID: symbol.ID, Time: marketdata.MustParseUTCTime("2026-03-01T00:00:20Z"), Bid: 1004, Ask: 1006},
		{SymbolID: symbol.ID, Time: marketdata.MustParseUTCTime("2026-03-01T00:01:05Z"), Bid: 1010, Ask: 1014},
	})
	if err != nil {
		t.Fatalf("upsert ticks: %v", err)
	}

	result, err := backfill.Execute(ctx, BackfillTimeframeBarsRequest{
		SymbolCode:    code,
		TimeframeCode: timeframe.TimeframeM1,
		From:          marketdata.MustParseUTCTime("2026-03-01T00:00:00Z"),
		To:            marketdata.MustParseUTCTime("2026-03-01T00:02:00Z"),
		ChunkSizeBars: 1,
	})
	if err != nil {
		t.Fatalf("backfill execute: %v", err)
	}
	if result.BarCount != 2 {
		t.Fatalf("expected 2 bars in result, got %d", result.BarCount)
	}

	items, err := listBars.Execute(
		ctx,
		symbol.ID,
		timeframe.TimeframeM1,
		marketdata.MustParseUTCTime("2026-03-01T00:00:00Z"),
		marketdata.MustParseUTCTime("2026-03-01T00:03:00Z"),
	)
	if err != nil {
		t.Fatalf("list timeframe bars: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 persisted bars, got %d", len(items))
	}
	if got := items[0].Open.Raw(); got != 1001 {
		t.Fatalf("expected first bar open=1001, got %d", got)
	}
	if got := items[0].Close.Raw(); got != 1005 {
		t.Fatalf("expected first bar close=1005, got %d", got)
	}
	if got := items[0].High.Raw(); got != 1006 {
		t.Fatalf("expected first bar high=1006, got %d", got)
	}
	if got := items[0].Low.Raw(); got != 1000 {
		t.Fatalf("expected first bar low=1000, got %d", got)
	}
	if !items[0].Hightime.Equal(marketdata.MustParseUTCTime("2026-03-01T00:00:20Z")) {
		t.Fatalf("expected first bar high_time=2026-03-01T00:00:20Z, got %s", items[0].Hightime)
	}
	if !items[0].Lowtime.Equal(marketdata.MustParseUTCTime("2026-03-01T00:00:10Z")) {
		t.Fatalf("expected first bar low_time=2026-03-01T00:00:10Z, got %s", items[0].Lowtime)
	}
	if got := int64(items[0].Volume); got != 2 {
		t.Fatalf("expected first bar volume=2, got %d", got)
	}
	if items[0].Source != timeframe.TimeframeBarSourceTickBidAskMid {
		t.Fatalf("expected first bar source=%q, got %q", timeframe.TimeframeBarSourceTickBidAskMid, items[0].Source)
	}
}

func cleanupBackfillIntegrationFixtures(ctx context.Context, t *testing.T, client *tsdb.Client, code string) {
	t.Helper()
	_, _ = client.Pool.Exec(ctx, `DELETE FROM timeframe_bar WHERE symbol_id IN (SELECT id FROM symbol WHERE code = $1)`, code)
	_, _ = client.Pool.Exec(ctx, `DELETE FROM tick WHERE symbol_id IN (SELECT id FROM symbol WHERE code = $1)`, code)
	_, _ = client.Pool.Exec(ctx, `DELETE FROM symbol WHERE code = $1`, code)
}
