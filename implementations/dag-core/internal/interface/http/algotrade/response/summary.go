package response

type StrategySummary struct {
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

type GetSummaryResponse struct {
	Partition string          `json:"partition"`
	Summary   StrategySummary `json:"summary"`
}
