package marketdata

import (
	"context"
	"os"
	"testing"
	"time"

	domainmarketdata "dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
)

func TestSessionBarRepositoryRoundTripToTSDB(t *testing.T) {
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
	code := "ITSESBARREPO"
	cleanupSessionBarFixtures(ctx, t, client, code)
	defer cleanupSessionBarFixtures(ctx, t, client, code)

	symbolRepo := NewSymbolRepository(client)
	barRepo := NewSessionBarRepository(client)

	symbol, err := symbolRepo.Create(ctx, createIntegrationSymbolInput(code))
	if err != nil {
		t.Fatalf("create symbol: %v", err)
	}

	bars := []domainmarketdata.SessionBar{
		{
			SymbolID:    symbol.ID,
			SessionCode: domainmarketdata.SessionTokyo,
			SessionDate: domainmarketdata.MustSessionDate(2026, time.March, 2),
			OHLCV: domainmarketdata.OHLCV{
				Opentime:  domainmarketdata.MustParseUTCTime("2026-03-02T00:00:00Z"),
				Closetime: domainmarketdata.MustParseUTCTime("2026-03-02T06:00:00Z"),
				Open:      domainmarketdata.NewPriceFromRaw(1001),
				High:      domainmarketdata.NewPriceFromRaw(1007),
				Hightime:  domainmarketdata.MustParseUTCTime("2026-03-02T02:15:00Z"),
				Low:       domainmarketdata.NewPriceFromRaw(999),
				Lowtime:   domainmarketdata.MustParseUTCTime("2026-03-02T00:30:00Z"),
				Close:     domainmarketdata.NewPriceFromRaw(1004),
				Volume:    3,
			},
			Source: domainmarketdata.TimeframeBarSourceTickBidAskMid,
		},
		{
			SymbolID:    symbol.ID,
			SessionCode: domainmarketdata.SessionTokyo,
			SessionDate: domainmarketdata.MustSessionDate(2026, time.March, 3),
			OHLCV: domainmarketdata.OHLCV{
				Opentime:  domainmarketdata.MustParseUTCTime("2026-03-03T00:00:00Z"),
				Closetime: domainmarketdata.MustParseUTCTime("2026-03-03T06:00:00Z"),
				Open:      domainmarketdata.NewPriceFromRaw(1005),
				High:      domainmarketdata.NewPriceFromRaw(1010),
				Hightime:  domainmarketdata.MustParseUTCTime("2026-03-03T05:15:00Z"),
				Low:       domainmarketdata.NewPriceFromRaw(1002),
				Lowtime:   domainmarketdata.MustParseUTCTime("2026-03-03T01:00:00Z"),
				Close:     domainmarketdata.NewPriceFromRaw(1008),
				Volume:    4,
			},
			Source: domainmarketdata.TimeframeBarSourceTickBidAskMid,
		},
	}

	if err := barRepo.BulkUpsert(ctx, bars); err != nil {
		t.Fatalf("BulkUpsert: %v", err)
	}

	items, err := barRepo.ListBySymbolSessionAndDateRange(
		ctx,
		symbol.ID,
		domainmarketdata.SessionTokyo,
		domainmarketdata.MustSessionDate(2026, time.March, 2),
		domainmarketdata.MustSessionDate(2026, time.March, 4),
	)
	if err != nil {
		t.Fatalf("ListBySymbolSessionAndDateRange: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(items))
	}
	if items[0].SessionDate.String() != "2026-03-02" {
		t.Fatalf("unexpected first session_date: %s", items[0].SessionDate.String())
	}
	if !items[0].Hightime.Equal(domainmarketdata.MustParseUTCTime("2026-03-02T02:15:00Z")) {
		t.Fatalf("unexpected first high_time: %s", items[0].Hightime)
	}

	latest, err := barRepo.GetLatestBySymbolAndSession(ctx, symbol.ID, domainmarketdata.SessionTokyo)
	if err != nil {
		t.Fatalf("GetLatestBySymbolAndSession: %v", err)
	}
	if latest.SessionDate.String() != "2026-03-03" {
		t.Fatalf("unexpected latest session_date: %s", latest.SessionDate.String())
	}
	if !latest.Lowtime.Equal(domainmarketdata.MustParseUTCTime("2026-03-03T01:00:00Z")) {
		t.Fatalf("unexpected latest low_time: %s", latest.Lowtime)
	}

	if err := barRepo.DeleteBySymbolSessionAndDateRange(
		ctx,
		symbol.ID,
		domainmarketdata.SessionTokyo,
		domainmarketdata.MustSessionDate(2026, time.March, 2),
		domainmarketdata.MustSessionDate(2026, time.March, 3),
	); err != nil {
		t.Fatalf("DeleteBySymbolSessionAndDateRange: %v", err)
	}

	items, err = barRepo.ListBySymbolSessionAndDateRange(
		ctx,
		symbol.ID,
		domainmarketdata.SessionTokyo,
		domainmarketdata.MustSessionDate(2026, time.March, 2),
		domainmarketdata.MustSessionDate(2026, time.March, 4),
	)
	if err != nil {
		t.Fatalf("ListBySymbolSessionAndDateRange after delete: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 remaining bar, got %d", len(items))
	}
	if items[0].SessionDate.String() != "2026-03-03" {
		t.Fatalf("unexpected remaining session_date: %s", items[0].SessionDate.String())
	}
}

func cleanupSessionBarFixtures(ctx context.Context, t *testing.T, client *tsdb.Client, code string) {
	t.Helper()
	_, _ = client.Pool.Exec(ctx, `DELETE FROM session_bar WHERE symbol_id IN (SELECT id FROM symbol WHERE code = $1)`, code)
	_, _ = client.Pool.Exec(ctx, `DELETE FROM timeframe_bar WHERE symbol_id IN (SELECT id FROM symbol WHERE code = $1)`, code)
	_, _ = client.Pool.Exec(ctx, `DELETE FROM tick WHERE symbol_id IN (SELECT id FROM symbol WHERE code = $1)`, code)
	_, _ = client.Pool.Exec(ctx, `DELETE FROM symbol WHERE code = $1`, code)
}
