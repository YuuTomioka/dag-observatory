package marketcontext

import (
	"testing"
	"time"

	"dag-observatory/dag-core/internal/domain/marketphase"
)

func TestResolveSignalsOverlap(t *testing.T) {
	t.Parallel()

	at := mustParseRFC3339("2026-03-20T12:30:00Z")
	singles, err := marketphase.ResolveSinglePhases(at, marketphase.DefaultPhaseSpecs())
	if err != nil {
		t.Fatalf("ResolveSinglePhases: %v", err)
	}
	signals := ResolveSignals(singles)
	if len(signals) != 1 || signals[0] != SignalLondonNewYorkOverlap {
		t.Fatalf("unexpected signals: %#v", signals)
	}
}

func TestResolveSignalsTokyoLondonTransition(t *testing.T) {
	t.Parallel()

	at := mustParseRFC3339("2026-07-15T06:30:00Z")
	singles, err := marketphase.ResolveSinglePhases(at, marketphase.DefaultPhaseSpecs())
	if err != nil {
		t.Fatalf("ResolveSinglePhases: %v", err)
	}
	assertSignalList(t, ResolveSignals(singles), []SignalID{SignalTokyoLondonTransition})
}

func TestBuildDeterministic(t *testing.T) {
	t.Parallel()

	at := mustParseRFC3339("2026-03-20T12:30:00Z")
	singles, err := marketphase.ResolveSinglePhases(at, marketphase.DefaultPhaseSpecs())
	if err != nil {
		t.Fatalf("ResolveSinglePhases: %v", err)
	}
	signals := ResolveSignals(singles)

	got1 := Build(at, singles, signals)
	got2 := Build(at, singles, signals)
	if !got1.NowUTC.Equal(got2.NowUTC) {
		t.Fatalf("NowUTC mismatch: %s vs %s", got1.NowUTC, got2.NowUTC)
	}
	assertPhaseList(t, got1.ActivePhases, got2.ActivePhases)
	assertSignalList(t, got1.ActiveSignals, got2.ActiveSignals)
	assertStringList(t, got1.DominantMarkets, got2.DominantMarkets)
	assertStringList(t, got1.Tags, got2.Tags)
	if len(got1.Tags) == 0 {
		t.Fatal("expected merged tags")
	}
}

func TestResolveIncludesTransition(t *testing.T) {
	t.Parallel()

	ctx, err := Resolve(mustParseRFC3339("2026-07-15T06:30:00Z"), marketphase.DefaultPhaseSpecs())
	if err != nil {
		t.Fatalf("Resolve transition: %v", err)
	}
	assertPhaseList(t, ctx.ActivePhases, []marketphase.PhaseCode{marketphase.PhaseLateTokyo})
	assertSignalList(t, ctx.ActiveSignals, []SignalID{SignalTokyoLondonTransition})
}

func assertSignalList(t *testing.T, got, want []SignalID) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("unexpected signal list length: got=%#v want=%#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected signal list: got=%#v want=%#v", got, want)
		}
	}
}

func assertStringList(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("unexpected string list length: got=%#v want=%#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected string list: got=%#v want=%#v", got, want)
		}
	}
}

func assertPhaseList(t *testing.T, got, want []marketphase.PhaseCode) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("unexpected phase list length: got=%#v want=%#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected phase list: got=%#v want=%#v", got, want)
		}
	}
}

func mustParseRFC3339(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
