package marketphase

import (
	"testing"
	"time"
)

func TestResolveSinglePhasesTable(t *testing.T) {
	t.Parallel()

	specs := DefaultPhaseSpecs()
	tests := []struct {
		name     string
		at       string
		expected []PhaseID
	}{
		{
			name:     "tokyo core normal day",
			at:       "2026-01-15T01:00:00Z",
			expected: []PhaseID{PhaseTokyoCore},
		},
		{
			name:     "london early during UK winter",
			at:       "2026-01-15T08:30:00Z",
			expected: []PhaseID{PhaseLondonEarly},
		},
		{
			name:     "london early during UK summer",
			at:       "2026-07-15T07:30:00Z",
			expected: []PhaseID{PhaseLondonEarly},
		},
		{
			name:     "new york core during US winter",
			at:       "2026-01-15T13:30:00Z",
			expected: []PhaseID{PhaseLondonLate, PhaseNewYorkCore},
		},
		{
			name:     "new york core during US summer",
			at:       "2026-07-15T12:30:00Z",
			expected: []PhaseID{PhaseLondonLate, PhaseNewYorkCore},
		},
		{
			name:     "us uk dst misaligned overlap period",
			at:       "2026-03-20T12:30:00Z",
			expected: []PhaseID{PhaseLondonCore, PhaseNewYorkCore},
		},
		{
			name:     "late tokyo before london early still active single only",
			at:       "2026-07-15T06:30:00Z",
			expected: []PhaseID{PhaseLateTokyo},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ResolveSinglePhases(mustParseRFC3339(tc.at), specs)
			if err != nil {
				t.Fatalf("ResolveSinglePhases: %v", err)
			}
			assertPhaseIDs(t, got, tc.expected)
		})
	}
}

func TestResolveDerivedPhasesOverlap(t *testing.T) {
	t.Parallel()

	at := mustParseRFC3339("2026-03-20T12:30:00Z")
	singles, err := ResolveSinglePhases(at, DefaultPhaseSpecs())
	if err != nil {
		t.Fatalf("ResolveSinglePhases: %v", err)
	}
	derived := ResolveDerivedPhases(singles)
	if len(derived) != 1 || derived[0] != PhaseLondonNewYorkOverlap {
		t.Fatalf("unexpected derived phases: %#v", derived)
	}
}

func TestResolveDerivedPhasesTokyoLondonTransition(t *testing.T) {
	t.Parallel()

	at := mustParseRFC3339("2026-07-15T06:30:00Z")
	singles, err := ResolveSinglePhases(at, DefaultPhaseSpecs())
	if err != nil {
		t.Fatalf("ResolveSinglePhases: %v", err)
	}
	derived := ResolveDerivedPhases(singles)
	assertPhaseIDList(t, derived, []PhaseID{PhaseTokyoLondonTransition})
}

func TestResolveSinglePhasesBoundaryHalfOpen(t *testing.T) {
	t.Parallel()

	specs := DefaultPhaseSpecs()
	atStart := mustParseRFC3339("2026-01-15T23:00:00Z")
	singles, err := ResolveSinglePhases(atStart, specs)
	if err != nil {
		t.Fatalf("ResolveSinglePhases at start: %v", err)
	}
	assertPhaseIDs(t, singles, []PhaseID{PhasePreTokyo})

	atEnd := mustParseRFC3339("2026-01-16T00:00:00Z")
	singles, err = ResolveSinglePhases(atEnd, specs)
	if err != nil {
		t.Fatalf("ResolveSinglePhases at end: %v", err)
	}
	assertPhaseIDs(t, singles, []PhaseID{PhaseTokyoCore})
}

func TestResolveSinglePhasesJSTWindowsShiftSeasonally(t *testing.T) {
	t.Parallel()

	specs := DefaultPhaseSpecs()
	winterSingles, err := ResolveSinglePhases(mustParseRFC3339("2026-01-15T08:30:00Z"), specs)
	if err != nil {
		t.Fatalf("winter resolve: %v", err)
	}
	summerSingles, err := ResolveSinglePhases(mustParseRFC3339("2026-07-15T07:30:00Z"), specs)
	if err != nil {
		t.Fatalf("summer resolve: %v", err)
	}

	winter := mustFindPhase(t, winterSingles, PhaseLondonEarly)
	summer := mustFindPhase(t, summerSingles, PhaseLondonEarly)
	if winter.JSTStart.Hour() != 17 || summer.JSTStart.Hour() != 16 {
		t.Fatalf("unexpected JST starts: winter=%s summer=%s", winter.JSTStart, summer.JSTStart)
	}
	if winter.LocalStart.Hour() != 8 || summer.LocalStart.Hour() != 8 {
		t.Fatalf("expected stable local starts at 08:00, got winter=%s summer=%s", winter.LocalStart, summer.LocalStart)
	}
}

func TestBuildMarketContextDeterministic(t *testing.T) {
	t.Parallel()

	at := mustParseRFC3339("2026-03-20T12:30:00Z")
	singles, err := ResolveSinglePhases(at, DefaultPhaseSpecs())
	if err != nil {
		t.Fatalf("ResolveSinglePhases: %v", err)
	}
	derived := ResolveDerivedPhases(singles)

	got1 := BuildMarketContext(at, singles, derived)
	got2 := BuildMarketContext(at, singles, derived)
	if !got1.NowUTC.Equal(got2.NowUTC) {
		t.Fatalf("NowUTC mismatch: %s vs %s", got1.NowUTC, got2.NowUTC)
	}
	assertPhaseIDList(t, got1.ActiveSinglePhases, got2.ActiveSinglePhases)
	assertPhaseIDList(t, got1.ActiveDerivedPhases, got2.ActiveDerivedPhases)
	assertStringList(t, got1.DominantMarkets, got2.DominantMarkets)
	assertStringList(t, got1.Tags, got2.Tags)
	if got1.LiquidityScore != got2.LiquidityScore || got1.BreakoutScore != got2.BreakoutScore || got1.MeanReversionScore != got2.MeanReversionScore {
		t.Fatalf("score mismatch: %#v vs %#v", got1, got2)
	}
	if got1.LiquidityScore != 1.0 || got1.BreakoutScore != 1.0 {
		t.Fatalf("expected overlap to raise liquidity/breakout to 1.0, got liquidity=%.2f breakout=%.2f", got1.LiquidityScore, got1.BreakoutScore)
	}
}

func TestResolveMarketContextIncludesTransition(t *testing.T) {
	t.Parallel()

	ctx, err := ResolveMarketContext(mustParseRFC3339("2026-07-15T06:30:00Z"), DefaultPhaseSpecs())
	if err != nil {
		t.Fatalf("ResolveMarketContext transition: %v", err)
	}
	assertPhaseIDList(t, ctx.ActiveSinglePhases, []PhaseID{PhaseLateTokyo})
	assertPhaseIDList(t, ctx.ActiveDerivedPhases, []PhaseID{PhaseTokyoLondonTransition})
}

func assertPhaseIDs(t *testing.T, phases []ResolvedPhase, want []PhaseID) {
	t.Helper()
	got := make([]PhaseID, 0, len(phases))
	for _, phase := range phases {
		got = append(got, phase.ID)
	}
	assertPhaseIDList(t, got, want)
}

func assertPhaseIDList(t *testing.T, got, want []PhaseID) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("unexpected phase ids length: got=%#v want=%#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected phase ids: got=%#v want=%#v", got, want)
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

func mustFindPhase(t *testing.T, phases []ResolvedPhase, id PhaseID) ResolvedPhase {
	t.Helper()
	for _, phase := range phases {
		if phase.ID == id {
			return phase
		}
	}
	t.Fatalf("phase not found: %s", id)
	return ResolvedPhase{}
}

func mustParseRFC3339(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
