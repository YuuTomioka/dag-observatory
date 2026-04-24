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
		expected []PhaseCode
	}{
		{
			name:     "tokyo core normal day",
			at:       "2026-01-15T01:00:00Z",
			expected: []PhaseCode{PhaseTokyoCore},
		},
		{
			name:     "london early during UK winter",
			at:       "2026-01-15T08:30:00Z",
			expected: []PhaseCode{PhaseLondonEarly},
		},
		{
			name:     "london early during UK summer",
			at:       "2026-07-15T07:30:00Z",
			expected: []PhaseCode{PhaseLondonEarly},
		},
		{
			name:     "new york core during US winter",
			at:       "2026-01-15T13:30:00Z",
			expected: []PhaseCode{PhaseLondonLate, PhaseNewYorkCore},
		},
		{
			name:     "new york core during US summer",
			at:       "2026-07-15T12:30:00Z",
			expected: []PhaseCode{PhaseLondonLate, PhaseNewYorkCore},
		},
		{
			name:     "us uk dst misaligned overlap period",
			at:       "2026-03-20T12:30:00Z",
			expected: []PhaseCode{PhaseLondonCore, PhaseNewYorkCore},
		},
		{
			name:     "late tokyo before london early still active single only",
			at:       "2026-07-15T06:30:00Z",
			expected: []PhaseCode{PhaseLateTokyo},
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
			assertPhaseCodes(t, got, tc.expected)
		})
	}
}

func TestResolveSinglePhasesBoundaryHalfOpen(t *testing.T) {
	t.Parallel()

	specs := DefaultPhaseSpecs()
	atStart := mustParseRFC3339("2026-01-15T23:00:00Z")
	singles, err := ResolveSinglePhases(atStart, specs)
	if err != nil {
		t.Fatalf("ResolveSinglePhases at start: %v", err)
	}
	assertPhaseCodes(t, singles, []PhaseCode{PhasePreTokyo})

	atEnd := mustParseRFC3339("2026-01-16T00:00:00Z")
	singles, err = ResolveSinglePhases(atEnd, specs)
	if err != nil {
		t.Fatalf("ResolveSinglePhases at end: %v", err)
	}
	assertPhaseCodes(t, singles, []PhaseCode{PhaseTokyoCore})
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

func assertPhaseCodes(t *testing.T, phases []ResolvedPhase, want []PhaseCode) {
	t.Helper()
	got := make([]PhaseCode, 0, len(phases))
	for _, phase := range phases {
		got = append(got, phase.ID)
	}
	assertPhaseCodeList(t, got, want)
}

func assertPhaseCodeList(t *testing.T, got, want []PhaseCode) {
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

func mustFindPhase(t *testing.T, phases []ResolvedPhase, id PhaseCode) ResolvedPhase {
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
