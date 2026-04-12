package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	apprepository "dag-observatory/dag-core/internal/application/dagruntime/repository"
)

type GetRunBacktestSummary struct {
	Reader      port.RunReader
	SummaryRepo apprepository.BacktestSummaryReader
	Writer      apprepository.BacktestSummaryWriter
}

type GetRunBacktestSummaryRequest struct {
	RunID string
}

type BacktestSummaryView struct {
	RunID             string  `json:"run_id"`
	Partition         string  `json:"partition"`
	TradeCount        int64   `json:"trade_count"`
	WinRate           float64 `json:"win_rate"`
	TotalNetPnL       float64 `json:"total_net_pnl"`
	MaxDrawdown       float64 `json:"max_drawdown"`
	EquityPointCount  int64   `json:"equity_point_count"`
	HasStrategyFields bool    `json:"has_strategy_fields"`
}

func (u *GetRunBacktestSummary) Execute(ctx context.Context, req GetRunBacktestSummaryRequest) (BacktestSummaryView, bool, error) {
	if u == nil {
		return BacktestSummaryView{}, false, nil
	}
	if req.RunID == "" {
		return BacktestSummaryView{}, false, fmt.Errorf("run_id is required")
	}
	if u.SummaryRepo != nil {
		item, err := u.SummaryRepo.GetByRunID(ctx, req.RunID)
		switch {
		case err == nil:
			return BacktestSummaryView{
				RunID:             item.RunID,
				Partition:         item.Partition,
				TradeCount:        item.TradeCount,
				WinRate:           item.WinRate,
				TotalNetPnL:       item.TotalNetPnL,
				MaxDrawdown:       item.MaxDrawdown,
				EquityPointCount:  item.EquityPointCount,
				HasStrategyFields: item.HasStrategyFields,
			}, true, nil
		case errors.Is(err, apprepository.ErrNotFound), errors.Is(err, apprepository.ErrNotConfigured):
			// Fall through to reconstruction from in-memory run steps.
		default:
			// Keep API available even when TSDB read path is temporarily unavailable.
			if u.Reader == nil {
				return BacktestSummaryView{}, false, err
			}
		}
	}
	if u.Reader == nil {
		return BacktestSummaryView{}, false, nil
	}
	run, ok := u.Reader.GetRun(req.RunID)
	if !ok {
		return BacktestSummaryView{}, false, nil
	}
	steps := u.Reader.ListRunSteps(req.RunID)

	out := BacktestSummaryView{
		RunID:     run.RunID,
		Partition: string(run.Partition),
	}
	for _, step := range steps {
		for _, diff := range step.StateDiff {
			switch diff.Field {
			case "strategy_summary.trade_count":
				if n, ok := toInt64(diff.After); ok {
					out.TradeCount = n
					out.HasStrategyFields = true
				}
			case "strategy_summary.win_rate":
				if n, ok := toFloat64(diff.After); ok {
					out.WinRate = n
					out.HasStrategyFields = true
				}
			case "strategy_summary.total_net_pnl":
				if n, ok := toFloat64(diff.After); ok {
					out.TotalNetPnL = n
					out.HasStrategyFields = true
				}
			case "strategy_summary.max_drawdown":
				if n, ok := toFloat64(diff.After); ok {
					out.MaxDrawdown = n
					out.HasStrategyFields = true
				}
			case "equity_series.count":
				if n, ok := toInt64(diff.After); ok {
					out.EquityPointCount = n
				}
			}
		}
	}
	if u.Writer != nil {
		_ = u.Writer.Upsert(ctx, apprepository.BacktestRunSummary{
			RunID:             out.RunID,
			Partition:         out.Partition,
			TradeCount:        out.TradeCount,
			WinRate:           out.WinRate,
			TotalNetPnL:       out.TotalNetPnL,
			MaxDrawdown:       out.MaxDrawdown,
			EquityPointCount:  out.EquityPointCount,
			HasStrategyFields: out.HasStrategyFields,
		})
	}
	return out, true, nil
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case string:
		parsed, err := strconv.ParseFloat(n, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func toInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int64:
		return n, true
	case int:
		return int64(n), true
	case int32:
		return int64(n), true
	case float64:
		return int64(n), true
	case float32:
		return int64(n), true
	case string:
		parsed, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
