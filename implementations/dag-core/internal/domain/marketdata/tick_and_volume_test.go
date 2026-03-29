package marketdata

import "testing"

// TestTickSpread は Tick のスプレッド計算（Ask - Bid）を検証する。
func TestTickSpread(t *testing.T) {
	tick := Tick{
		SymbolID: 1,
		Bid:      NewPriceFromRaw(100),
		Ask:      NewPriceFromRaw(104),
	}
	if got := tick.Spread().Raw(); got != 4 {
		t.Fatalf("Spread mismatch: got=%d want=4", got)
	}
}

// TestVolumeAdd は Volume の加算結果を検証する。
func TestVolumeAdd(t *testing.T) {
	if got := Volume(7).Add(Volume(5)); got != 12 {
		t.Fatalf("Volume.Add mismatch: got=%d want=12", got)
	}
}
