package marketphase

import (
	"testing"
	"time"
)

func TestParsePhaseDate(t *testing.T) {
	t.Parallel()

	d, err := ParsePhaseDate("2026-07-15")
	if err != nil {
		t.Fatalf("ParsePhaseDate: %v", err)
	}
	if d.String() != "2026-07-15" {
		t.Fatalf("unexpected phase date string: %s", d.String())
	}
}

func TestNewPhaseDateFromTime(t *testing.T) {
	t.Parallel()

	d := NewPhaseDateFromTime(time.Date(2026, 7, 15, 23, 59, 0, 0, time.FixedZone("EDT", -4*60*60)))
	if d.String() != "2026-07-15" {
		t.Fatalf("unexpected phase date: %s", d.String())
	}
}

func TestNewPhaseDateFromOpentimeUTC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "before cutover",
			in:   "2026-07-15T21:59:59Z",
			want: "2026-07-15",
		},
		{
			name: "at cutover",
			in:   "2026-07-15T22:00:00Z",
			want: "2026-07-16",
		},
		{
			name: "after cutover",
			in:   "2026-07-15T23:30:00Z",
			want: "2026-07-16",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := NewPhaseDateFromOpentimeUTC(mustParseRFC3339PhaseDate(tc.in))
			if got.String() != tc.want {
				t.Fatalf("phase date mismatch: got=%s want=%s", got.String(), tc.want)
			}
		})
	}
}

func mustParseRFC3339PhaseDate(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
