package algotrade

import "testing"

func TestStateKeysStableIDs(t *testing.T) {
	t.Parallel()

	if got := StateOpenPositions.Raw().StableID; got != "state:algotrade.open_positions.v1" {
		t.Fatalf("unexpected open positions stable id: %s", got)
	}
	if got := StatePendingOrders.Raw().StableID; got != "state:algotrade.pending_orders.v1" {
		t.Fatalf("unexpected pending orders stable id: %s", got)
	}
	if got := StateDailyPnL.Raw().StableID; got != "state:algotrade.daily_pnl.v1" {
		t.Fatalf("unexpected daily pnl stable id: %s", got)
	}
	if got := StateLossStreak.Raw().StableID; got != "state:algotrade.loss_streak.v1" {
		t.Fatalf("unexpected loss streak stable id: %s", got)
	}
}
