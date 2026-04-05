package algotrade

import (
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

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
	StateClosedTrades = state.Key[ClosedTradesState]{
		Name:     "closed_trades",
		StableID: "state:algotrade.closed_trades.v1",
	}
	StateStrategySummary = state.Key[StrategySummaryState]{
		Name:     "strategy_summary",
		StableID: "state:algotrade.strategy_summary.v1",
	}
	StateLossStreak = state.Key[LossStreakState]{
		Name:     "loss_streak",
		StableID: "state:algotrade.loss_streak.v1",
	}
)

type OpenPosition struct {
	PositionID      string
	IntentID        string
	ExecutionID     string
	Symbol          string
	Side            PositionSide
	Size            float64
	EntryPrice      marketdata.Price
	EntryTime       marketdata.UTCTime
	EntryReason     string
	StrategyID      string
	WorkflowName    string
	WorkflowVersion string
	ParameterSetID  string
	Snapshot        PositionSnapshot
}

type OpenPositionsState struct {
	Items []OpenPosition
}

type PendingOrder struct {
	OrderID     string
	IntentID    string
	ExecutionID string
	Symbol      string
	Intent      TradeIntent
	Execution   ExecutionResult
	SourceID    string
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

type ClosedTradesState struct {
	Items []ClosedTrade
}

type StrategySummaryState struct {
	Summary StrategySummary
}

type LossStreakState struct {
	ConsecutiveLosingTrades int
	ConsecutiveLosingDays   int
	LastLossDay             string
}

func (s DailyPnLState) ToSnapshot() PnLSnapshot {
	return PnLSnapshot{
		TradingDay:    s.TradingDay,
		RealizedPnL:   s.RealizedPnL,
		UnrealizedPnL: s.UnrealizedPnL,
		LossLimitHit:  s.LossLimitHit,
	}
}

func (p OpenPosition) SnapshotView() PositionSnapshot {
	if p.Snapshot.HasPosition || p.Snapshot.Side != "" || p.Snapshot.Size > 0 || p.Snapshot.EntryPrice.Raw() > 0 {
		return p.Snapshot
	}
	if p.PositionID == "" && p.Symbol == "" && p.Size == 0 {
		return PositionSnapshot{
			HasPosition: false,
			Side:        PositionSideFlat,
			Size:        0,
			EntryPrice:  marketdata.Price(0),
		}
	}
	return PositionSnapshot{
		HasPosition: p.Side != PositionSideFlat && p.Size > 0,
		Side:        p.Side,
		Size:        p.Size,
		EntryPrice:  p.EntryPrice,
	}
}
