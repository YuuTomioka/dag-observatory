package marketphase

import (
	"testing"

	"dag-observatory/dag-core/internal/domain/marketdata"
)

func TestAggregatePhaseBarUsesResolvedSingleWindow(t *testing.T) {
	t.Parallel()

	phase := mustFindPhase(t, mustResolveSingles(t, "2026-07-15T12:30:00Z"), PhaseNewYorkCore)
	bar, ok, err := AggregatePhaseBar(phase, 1, []marketdata.Tick{
		{
			SymbolID: 1,
			Time:     marketdata.MustParseUTCTime("2026-07-15T12:00:00Z"),
			Bid:      marketdata.NewPriceFromRaw(1000),
			Ask:      marketdata.NewPriceFromRaw(1002),
		},
		{
			SymbolID: 1,
			Time:     marketdata.MustParseUTCTime("2026-07-15T12:15:00Z"),
			Bid:      marketdata.NewPriceFromRaw(1001),
			Ask:      marketdata.NewPriceFromRaw(1004),
		},
		{
			SymbolID: 1,
			Time:     marketdata.MustParseUTCTime("2026-07-15T14:59:00Z"),
			Bid:      marketdata.NewPriceFromRaw(999),
			Ask:      marketdata.NewPriceFromRaw(1003),
		},
		{
			SymbolID: 1,
			Time:     marketdata.MustParseUTCTime("2026-07-15T15:00:00Z"),
			Bid:      marketdata.NewPriceFromRaw(1005),
			Ask:      marketdata.NewPriceFromRaw(1007),
		},
	})
	if err != nil {
		t.Fatalf("AggregatePhaseBar: %v", err)
	}
	if !ok {
		t.Fatal("expected phase bar")
	}
	if bar.PhaseID != PhaseNewYorkCore {
		t.Fatalf("phase id mismatch: got=%q", bar.PhaseID)
	}
	if bar.Market != "New York" || bar.Timezone != "America/New_York" {
		t.Fatalf("phase metadata mismatch: market=%q timezone=%q", bar.Market, bar.Timezone)
	}
	if !bar.Opentime.Equal(marketdata.MustParseUTCTime("2026-07-15T12:00:00Z")) {
		t.Fatalf("unexpected open_time: %s", bar.Opentime)
	}
	if !bar.Closetime.Equal(marketdata.MustParseUTCTime("2026-07-15T15:00:00Z")) {
		t.Fatalf("unexpected close_time: %s", bar.Closetime)
	}
	if bar.Open.Raw() != 1001 || bar.Close.Raw() != 1001 {
		t.Fatalf("unexpected open/close: open=%d close=%d", bar.Open.Raw(), bar.Close.Raw())
	}
	if bar.High.Raw() != 1004 || !bar.Hightime.Equal(marketdata.MustParseUTCTime("2026-07-15T12:15:00Z")) {
		t.Fatalf("unexpected high/high_time: high=%d high_time=%s", bar.High.Raw(), bar.Hightime)
	}
	if bar.Low.Raw() != 999 || !bar.Lowtime.Equal(marketdata.MustParseUTCTime("2026-07-15T14:59:00Z")) {
		t.Fatalf("unexpected low/low_time: low=%d low_time=%s", bar.Low.Raw(), bar.Lowtime)
	}
	if bar.Volume != 3 {
		t.Fatalf("unexpected volume: %d", bar.Volume)
	}
}

func TestAggregatePhaseBarReturnsFalseWithoutTicksInWindow(t *testing.T) {
	t.Parallel()

	phase := mustFindPhase(t, mustResolveSingles(t, "2026-01-15T08:30:00Z"), PhaseLondonEarly)
	_, ok, err := AggregatePhaseBar(phase, 1, []marketdata.Tick{
		{
			SymbolID: 1,
			Time:     marketdata.MustParseUTCTime("2026-01-15T07:59:00Z"),
			Bid:      marketdata.NewPriceFromRaw(1000),
			Ask:      marketdata.NewPriceFromRaw(1001),
		},
	})
	if err != nil {
		t.Fatalf("AggregatePhaseBar: %v", err)
	}
	if ok {
		t.Fatal("expected no phase bar")
	}
}

func mustResolveSingles(t *testing.T, at string) []ResolvedPhase {
	t.Helper()
	singles, err := ResolveSinglePhases(mustParseRFC3339(at), DefaultPhaseSpecs())
	if err != nil {
		t.Fatalf("ResolveSinglePhases: %v", err)
	}
	return singles
}
