package algotrade

import "dag-observatory/dag-core/internal/domain/dagruntime/state"

// Partition scope is strategy_id + symbol (v0 assumption).

var (
	StateOpenPositions = state.Key[OpenPositionsState]{
		Name:     "open_positions",
		StableID: "state:algotrade.open_positions.v1",
	}
	StatePendingOrders = state.Key[PendingOrdersState]{
		Name:     "pending_orders",
		StableID: "state:algotrade.pending_orders.v1",
	}
	StateDailyPnL = state.Key[DailyPnLState]{
		Name:     "daily_pnl",
		StableID: "state:algotrade.daily_pnl.v1",
	}
	StateLossStreak = state.Key[LossStreakState]{
		Name:     "loss_streak",
		StableID: "state:algotrade.loss_streak.v1",
	}
)

type OpenPosition struct {
	PositionID string
	Symbol     string
	Snapshot   PositionSnapshot
}

type OpenPositionsState struct {
	Items []OpenPosition
}

type PendingOrder struct {
	OrderID  string
	Symbol   string
	Request  OrderRequest
	SourceID string
}

type PendingOrdersState struct {
	Items []PendingOrder
}

type DailyPnLState struct {
	TradingDay    string
	RealizedPnL   float64
	UnrealizedPnL float64
	LossLimitHit  bool
}

type LossStreakState struct {
	ConsecutiveLosingTrades int
	ConsecutiveLosingDays   int
	LastLossDay             string
}
