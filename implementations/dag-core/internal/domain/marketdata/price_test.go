package marketdata

import "testing"

// TestPriceArithmeticAndComparison は Price の基本演算と比較ヘルパーを検証する。
func TestPriceArithmeticAndComparison(t *testing.T) {
	p := NewPriceFromRaw(100)
	q := NewPriceFromRaw(35)

	if got := p.Raw(); got != 100 {
		t.Fatalf("Raw mismatch: got=%d want=100", got)
	}
	if got := p.Add(q).Raw(); got != 135 {
		t.Fatalf("Add mismatch: got=%d want=135", got)
	}
	if got := p.Sub(q).Raw(); got != 65 {
		t.Fatalf("Sub mismatch: got=%d want=65", got)
	}
	if got := q.Neg().Raw(); got != -35 {
		t.Fatalf("Neg mismatch: got=%d want=-35", got)
	}
	if got := NewPriceFromRaw(-9).Abs().Raw(); got != 9 {
		t.Fatalf("Abs mismatch: got=%d want=9", got)
	}
	if got := q.MulInt(3).Raw(); got != 105 {
		t.Fatalf("MulInt mismatch: got=%d want=105", got)
	}
	if got := p.DivInt(4).Raw(); got != 25 {
		t.Fatalf("DivInt mismatch: got=%d want=25", got)
	}
	if got := p.Cmp(q); got != 1 {
		t.Fatalf("Cmp mismatch: got=%d want=1", got)
	}
	if !p.Gt(q) || !p.Ge(q) || p.Lt(q) || p.Le(q) || !p.Ne(q) || p.Eq(q) {
		t.Fatalf("comparison helpers returned unexpected result")
	}
}

// TestPriceDivIntPanicsOnZero は 0 除算時に panic することを検証する。
func TestPriceDivIntPanicsOnZero(t *testing.T) {
	assertPanics(t, func() {
		_ = NewPriceFromRaw(1).DivInt(0)
	})
}
