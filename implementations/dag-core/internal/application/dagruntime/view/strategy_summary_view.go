package view

import "dag-observatory/dag-core/internal/domain/algotrade"

// StrategySummaryView is a presentation-oriented projection of a strategy summary.
type StrategySummaryView struct {
	StrategyID      string  `json:"strategy_id"`
	WorkflowName    string  `json:"workflow_name"`
	WorkflowVersion string  `json:"workflow_version"`
	ParameterSetID  string  `json:"parameter_set_id"`
	TradeCount      int     `json:"trade_count"`
	WinCount        int     `json:"win_count"`
	LossCount       int     `json:"loss_count"`
	WinRate         float64 `json:"win_rate"`
	TotalNetPnL     float64 `json:"total_net_pnl"`
	AverageWin      float64 `json:"average_win"`
	AverageLoss     float64 `json:"average_loss"`
	ProfitFactor    float64 `json:"profit_factor"`
	MaxDrawdown     float64 `json:"max_drawdown"`
	UpdatedAt       string  `json:"updated_at"`
}

func StrategySummaryFromDomain(summary algotrade.StrategySummary) StrategySummaryView {
	view := StrategySummaryView{
		StrategyID:      summary.StrategyID,
		WorkflowName:    summary.WorkflowName,
		WorkflowVersion: summary.WorkflowVersion,
		ParameterSetID:  summary.ParameterSetID,
		TradeCount:      summary.TradeCount,
		WinCount:        summary.WinCount,
		LossCount:       summary.LossCount,
		WinRate:         summary.WinRate,
		TotalNetPnL:     summary.TotalNetPnL,
		AverageWin:      summary.AverageWin,
		AverageLoss:     summary.AverageLoss,
		ProfitFactor:    summary.ProfitFactor,
		MaxDrawdown:     summary.MaxDrawdown,
	}
	if !summary.UpdatedAt.IsZero() {
		view.UpdatedAt = summary.UpdatedAt.Time().UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	return view
}
