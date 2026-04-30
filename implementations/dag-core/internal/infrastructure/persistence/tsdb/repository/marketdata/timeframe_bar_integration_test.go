package marketdata

import (
	"context"
	"os"
	"testing"

	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	domainmarketdata "dag-observatory/dag-core/internal/domain/marketdata"
	domainohlc "dag-observatory/dag-core/internal/domain/marketdata/ohlc"
	domaintimeframe "dag-observatory/dag-core/internal/domain/marketdata/ohlc/timeframe"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
)

func TestTimeframeBarRepositoryRoundTripToTSDB(t *testing.T) {
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
	code := "ITTFBARREPO"
	cleanupTimeframeBarFixtures(ctx, t, client, code)
	defer cleanupTimeframeBarFixtures(ctx, t, client, code)

	symbolRepo := NewSymbolRepository(client)
	barRepo := NewTimeframeBarRepository(client)

	symbol, err := symbolRepo.Create(ctx, createIntegrationSymbolInput(code))
	if err != nil {
		t.Fatalf("create symbol: %v", err)
	}

	bars := []domaintimeframe.TimeframeBar{
		{
			SymbolID:      symbol.ID,
			TimeframeCode: domaintimeframe.TimeframeM1,
			OHLCV: domainohlc.OHLCV{
				Opentime:  domainmarketdata.MustParseUTCTime("2026-03-01T00:00:00Z"),
				Closetime: domainmarketdata.MustParseUTCTime("2026-03-01T00:01:00Z"),
				Open:      domainmarketdata.NewPriceFromRaw(1001),
				High:      domainmarketdata.NewPriceFromRaw(1005),
				Hightime:  domainmarketdata.MustParseUTCTime("2026-03-01T00:00:30Z"),
				Low:       domainmarketdata.NewPriceFromRaw(1001),
				Lowtime:   domainmarketdata.MustParseUTCTime("2026-03-01T00:00:00Z"),
				Close:     domainmarketdata.NewPriceFromRaw(1004),
				Volume:    2,
			},
			Source: domaintimeframe.TimeframeBarSourceTickBidAskMid,
		},
		{
			SymbolID:      symbol.ID,
			TimeframeCode: domaintimeframe.TimeframeM1,
			OHLCV: domainohlc.OHLCV{
				Opentime:  domainmarketdata.MustParseUTCTime("2026-03-01T00:01:00Z"),
				Closetime: domainmarketdata.MustParseUTCTime("2026-03-01T00:02:00Z"),
				Open:      domainmarketdata.NewPriceFromRaw(1008),
				High:      domainmarketdata.NewPriceFromRaw(1012),
				Hightime:  domainmarketdata.MustParseUTCTime("2026-03-01T00:01:45Z"),
				Low:       domainmarketdata.NewPriceFromRaw(1008),
				Lowtime:   domainmarketdata.MustParseUTCTime("2026-03-01T00:01:00Z"),
				Close:     domainmarketdata.NewPriceFromRaw(1012),
				Volume:    1,
			},
			Source: domaintimeframe.TimeframeBarSourceTickBidAskMid,
		},
	}

	if err := barRepo.BulkUpsert(ctx, bars); err != nil {
		t.Fatalf("BulkUpsert: %v", err)
	}

	items, err := barRepo.ListBySymbolTimeframeAndRange(
		ctx,
		symbol.ID,
		domaintimeframe.TimeframeM1,
		domainmarketdata.MustParseUTCTime("2026-03-01T00:00:00Z"),
		domainmarketdata.MustParseUTCTime("2026-03-01T00:03:00Z"),
	)
	if err != nil {
		t.Fatalf("ListBySymbolTimeframeAndRange: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(items))
	}
	if !items[0].Hightime.Equal(domainmarketdata.MustParseUTCTime("2026-03-01T00:00:30Z")) {
		t.Fatalf("unexpected first high_time: %s", items[0].Hightime)
	}
	if !items[0].Lowtime.Equal(domainmarketdata.MustParseUTCTime("2026-03-01T00:00:00Z")) {
		t.Fatalf("unexpected first low_time: %s", items[0].Lowtime)
	}

	latest, err := barRepo.GetLatestBySymbolAndTimeframe(ctx, symbol.ID, domaintimeframe.TimeframeM1)
	if err != nil {
		t.Fatalf("GetLatestBySymbolAndTimeframe: %v", err)
	}
	if !latest.Opentime.Equal(domainmarketdata.MustParseUTCTime("2026-03-01T00:01:00Z")) {
		t.Fatalf("unexpected latest open_time: %s", latest.Opentime)
	}
	if !latest.Hightime.Equal(domainmarketdata.MustParseUTCTime("2026-03-01T00:01:45Z")) {
		t.Fatalf("unexpected latest high_time: %s", latest.Hightime)
	}
	if !latest.Lowtime.Equal(domainmarketdata.MustParseUTCTime("2026-03-01T00:01:00Z")) {
		t.Fatalf("unexpected latest low_time: %s", latest.Lowtime)
	}

	if err := barRepo.DeleteBySymbolTimeframeAndRange(
		ctx,
		symbol.ID,
		domaintimeframe.TimeframeM1,
		domainmarketdata.MustParseUTCTime("2026-03-01T00:00:00Z"),
		domainmarketdata.MustParseUTCTime("2026-03-01T00:01:00Z"),
	); err != nil {
		t.Fatalf("DeleteBySymbolTimeframeAndRange: %v", err)
	}

	items, err = barRepo.ListBySymbolTimeframeAndRange(
		ctx,
		symbol.ID,
		domaintimeframe.TimeframeM1,
		domainmarketdata.MustParseUTCTime("2026-03-01T00:00:00Z"),
		domainmarketdata.MustParseUTCTime("2026-03-01T00:03:00Z"),
	)
	if err != nil {
		t.Fatalf("ListBySymbolTimeframeAndRange after delete: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 remaining bar, got %d", len(items))
	}
}

func cleanupTimeframeBarFixtures(ctx context.Context, t *testing.T, client *tsdb.Client, code string) {
	t.Helper()
	_, _ = client.Pool.Exec(ctx, `DELETE FROM timeframe_bar WHERE symbol_id IN (SELECT id FROM symbol WHERE code = $1)`, code)
	_, _ = client.Pool.Exec(ctx, `DELETE FROM tick WHERE symbol_id IN (SELECT id FROM symbol WHERE code = $1)`, code)
	_, _ = client.Pool.Exec(ctx, `DELETE FROM symbol WHERE code = $1`, code)
}

func createIntegrationSymbolInput(code string) apprepository.CreateSymbolInput {
	return apprepository.CreateSymbolInput{
		Code:        code,
		Base:        "ITF",
		Quote:       "JPY",
		PriceScale:  5,
		TickSizeRaw: 1,
		PipSizeRaw:  10,
	}
}
