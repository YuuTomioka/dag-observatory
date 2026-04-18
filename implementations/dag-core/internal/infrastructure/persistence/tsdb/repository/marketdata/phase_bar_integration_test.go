package marketdata

import (
	"context"
	"os"
	"testing"

	domainmarketdata "dag-observatory/dag-core/internal/domain/marketdata"
	domainmarketphase "dag-observatory/dag-core/internal/domain/marketphase"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
)

func TestPhaseBarRepositoryRoundTripToTSDB(t *testing.T) {
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
	code := "ITPHBARREPO"
	cleanupPhaseBarFixtures(ctx, t, client, code)
	defer cleanupPhaseBarFixtures(ctx, t, client, code)

	symbolRepo := NewSymbolRepository(client)
	barRepo := NewPhaseBarRepository(client)
	symbol, err := symbolRepo.Create(ctx, createIntegrationSymbolInput(code))
	if err != nil {
		t.Fatalf("create symbol: %v", err)
	}

	bars := []domainmarketphase.PhaseBar{
		{
			PhaseID:  domainmarketphase.PhaseTokyoCore,
			Market:   "Tokyo",
			Timezone: "Asia/Tokyo",
			SymbolID: symbol.ID,
			OHLCV:    sampleOHLCV("2026-04-01T00:00:00Z", "2026-04-01T06:00:00Z", 1001, 1008, "2026-04-01T03:00:00Z", 999, "2026-04-01T00:30:00Z", 1005, 3),
			Source:   domainmarketphase.PhaseBarSourceTickBidAskMid,
		},
		{
			PhaseID:  domainmarketphase.PhaseTokyoCore,
			Market:   "Tokyo",
			Timezone: "Asia/Tokyo",
			SymbolID: symbol.ID,
			OHLCV:    sampleOHLCV("2026-04-02T00:00:00Z", "2026-04-02T06:00:00Z", 1004, 1010, "2026-04-02T05:00:00Z", 1002, "2026-04-02T01:00:00Z", 1007, 4),
			Source:   domainmarketphase.PhaseBarSourceTickBidAskMid,
		},
	}

	if err := barRepo.BulkUpsert(ctx, bars); err != nil {
		t.Fatalf("BulkUpsert: %v", err)
	}
	items, err := barRepo.ListBySymbolPhaseAndRange(ctx, symbol.ID, domainmarketphase.PhaseTokyoCore, mustUTC("2026-04-01T00:00:00Z"), mustUTC("2026-04-03T00:00:00Z"))
	if err != nil {
		t.Fatalf("ListBySymbolPhaseAndRange: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(items))
	}
	if items[0].Market != "Tokyo" || items[0].Timezone != "Asia/Tokyo" {
		t.Fatalf("unexpected metadata: market=%q timezone=%q", items[0].Market, items[0].Timezone)
	}
	latest, err := barRepo.GetLatestBySymbolAndPhase(ctx, symbol.ID, domainmarketphase.PhaseTokyoCore)
	if err != nil {
		t.Fatalf("GetLatestBySymbolAndPhase: %v", err)
	}
	if !latest.Opentime.Equal(mustUTC("2026-04-02T00:00:00Z")) {
		t.Fatalf("unexpected latest open_time: %s", latest.Opentime)
	}
	if err := barRepo.DeleteBySymbolPhaseAndRange(ctx, symbol.ID, domainmarketphase.PhaseTokyoCore, mustUTC("2026-04-01T00:00:00Z"), mustUTC("2026-04-02T00:00:00Z")); err != nil {
		t.Fatalf("DeleteBySymbolPhaseAndRange: %v", err)
	}
	items, err = barRepo.ListBySymbolPhaseAndRange(ctx, symbol.ID, domainmarketphase.PhaseTokyoCore, mustUTC("2026-04-01T00:00:00Z"), mustUTC("2026-04-03T00:00:00Z"))
	if err != nil {
		t.Fatalf("ListBySymbolPhaseAndRange after delete: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 remaining bar, got %d", len(items))
	}
}

func cleanupPhaseBarFixtures(ctx context.Context, t *testing.T, client *tsdb.Client, code string) {
	t.Helper()
	_, _ = client.Pool.Exec(ctx, `DELETE FROM phase_bar WHERE symbol_id IN (SELECT id FROM symbol WHERE code = $1)`, code)
	_, _ = client.Pool.Exec(ctx, `DELETE FROM timeframe_bar WHERE symbol_id IN (SELECT id FROM symbol WHERE code = $1)`, code)
	_, _ = client.Pool.Exec(ctx, `DELETE FROM tick WHERE symbol_id IN (SELECT id FROM symbol WHERE code = $1)`, code)
	_, _ = client.Pool.Exec(ctx, `DELETE FROM symbol WHERE code = $1`, code)
}

func sampleOHLCV(
	openTime string,
	closeTime string,
	open int64,
	high int64,
	highTime string,
	low int64,
	lowTime string,
	close int64,
	volume int64,
) domainmarketdata.OHLCV {
	return domainmarketdata.OHLCV{
		Opentime:  mustUTC(openTime),
		Closetime: mustUTC(closeTime),
		Open:      domainmarketdata.NewPriceFromRaw(open),
		High:      domainmarketdata.NewPriceFromRaw(high),
		Hightime:  mustUTC(highTime),
		Low:       domainmarketdata.NewPriceFromRaw(low),
		Lowtime:   mustUTC(lowTime),
		Close:     domainmarketdata.NewPriceFromRaw(close),
		Volume:    domainmarketdata.Volume(volume),
	}
}

func mustUTC(s string) domainmarketdata.UTCTime {
	return domainmarketdata.MustParseUTCTime(s)
}
