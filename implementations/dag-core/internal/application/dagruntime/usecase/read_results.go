package usecase

import (
	"context"
	"slices"
	"strings"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/view"
	"dag-observatory/dag-core/internal/domain/algotrade"
)

type ListTradeResults struct{}

type ListTradeResultsRequest struct {
	ClosedTrades algotrade.ClosedTradesState
	From         *time.Time
	To           *time.Time
	Symbol       string
	Side         string
	Limit        int
	Offset       int
}

func (u *ListTradeResults) Execute(ctx context.Context, req ListTradeResultsRequest) ([]view.TradeView, error) {
	_ = ctx
	trades := slices.Clone(req.ClosedTrades.Items)
	slices.SortFunc(trades, func(a, b algotrade.ClosedTrade) int {
		switch {
		case a.ExitTime.After(b.ExitTime):
			return -1
		case b.ExitTime.After(a.ExitTime):
			return 1
		default:
			if a.TradeID < b.TradeID {
				return -1
			}
			if a.TradeID > b.TradeID {
				return 1
			}
			return 0
		}
	})
	filtered := make([]algotrade.ClosedTrade, 0, len(trades))
	for _, trade := range trades {
		if !matchTradeFilter(trade, req) {
			continue
		}
		filtered = append(filtered, trade)
	}

	start := req.Offset
	if start < 0 {
		start = 0
	}
	if start >= len(filtered) {
		return []view.TradeView{}, nil
	}
	end := len(filtered)
	if req.Limit > 0 && start+req.Limit < end {
		end = start + req.Limit
	}

	items := make([]view.TradeView, 0, end-start)
	for _, trade := range filtered[start:end] {
		items = append(items, view.TradeFromClosedTrade(trade))
	}
	return items, nil
}

func matchTradeFilter(trade algotrade.ClosedTrade, req ListTradeResultsRequest) bool {
	if req.Symbol != "" && !strings.EqualFold(trade.Symbol, req.Symbol) {
		return false
	}
	if req.Side != "" && !strings.EqualFold(string(trade.Side), req.Side) {
		return false
	}
	if req.From != nil && trade.ExitTime.Time().Before(*req.From) {
		return false
	}
	if req.To != nil && !trade.ExitTime.Time().Before(*req.To) {
		return false
	}
	return true
}

type GetEquitySeries struct{}

type GetEquitySeriesRequest struct {
	ClosedTrades algotrade.ClosedTradesState
	From         *time.Time
	To           *time.Time
	Symbol       string
	Side         string
}

type EquityPointView struct {
	Time     string  `json:"time"`
	TradeID  string  `json:"trade_id"`
	Equity   float64 `json:"equity"`
	Drawdown float64 `json:"drawdown"`
}

type EquitySeriesView struct {
	Points      []EquityPointView `json:"points"`
	MaxDrawdown float64           `json:"max_drawdown"`
	TotalNetPnL float64           `json:"total_net_pnl"`
}

func (u *GetEquitySeries) Execute(ctx context.Context, req GetEquitySeriesRequest) (EquitySeriesView, error) {
	_ = ctx
	trades := slices.Clone(req.ClosedTrades.Items)
	slices.SortFunc(trades, func(a, b algotrade.ClosedTrade) int {
		switch {
		case a.ExitTime.Before(b.ExitTime):
			return -1
		case b.ExitTime.Before(a.ExitTime):
			return 1
		default:
			if a.TradeID < b.TradeID {
				return -1
			}
			if a.TradeID > b.TradeID {
				return 1
			}
			return 0
		}
	})

	out := EquitySeriesView{
		Points: make([]EquityPointView, 0, len(trades)),
	}
	var equity float64
	var peak float64
	for _, trade := range trades {
		if !matchTradeFilter(trade, ListTradeResultsRequest{
			From:   req.From,
			To:     req.To,
			Symbol: req.Symbol,
			Side:   req.Side,
		}) {
			continue
		}
		equity += trade.NetPnL
		if equity > peak {
			peak = equity
		}
		drawdown := peak - equity
		if drawdown > out.MaxDrawdown {
			out.MaxDrawdown = drawdown
		}
		out.Points = append(out.Points, EquityPointView{
			Time:     trade.ExitTime.Time().UTC().Format(time.RFC3339Nano),
			TradeID:  trade.TradeID,
			Equity:   equity,
			Drawdown: drawdown,
		})
	}
	out.TotalNetPnL = equity
	return out, nil
}

type GetStrategySummary struct{}

type GetStrategySummaryRequest struct {
	Summary algotrade.StrategySummaryState
}

func (u *GetStrategySummary) Execute(ctx context.Context, req GetStrategySummaryRequest) (view.StrategySummaryView, error) {
	_ = ctx
	return view.StrategySummaryFromDomain(req.Summary.Summary), nil
}
