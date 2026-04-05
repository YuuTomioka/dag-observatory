package view

import "dag-observatory/dag-core/internal/domain/algotrade"

// TradeView is a presentation-oriented projection of a closed trade fact.
type TradeView struct {
	TradeID         string  `json:"trade_id"`
	IntentID        string  `json:"intent_id"`
	PositionID      string  `json:"position_id"`
	Symbol          string  `json:"symbol"`
	Side            string  `json:"side"`
	Size            float64 `json:"size"`
	EntryTime       string  `json:"entry_time"`
	ExitTime        string  `json:"exit_time"`
	EntryPriceRaw   int64   `json:"entry_price_raw"`
	ExitPriceRaw    int64   `json:"exit_price_raw"`
	NetPnL          float64 `json:"net_pnl"`
	ExitReason      string  `json:"exit_reason"`
	WorkflowName    string  `json:"workflow_name"`
	WorkflowVersion string  `json:"workflow_version"`
	ParameterSetID  string  `json:"parameter_set_id"`
}

func TradeFromClosedTrade(trade algotrade.ClosedTrade) TradeView {
	view := TradeView{
		TradeID:         trade.TradeID,
		IntentID:        trade.IntentID,
		PositionID:      trade.PositionID,
		Symbol:          trade.Symbol,
		Side:            string(trade.Side),
		Size:            trade.Size,
		EntryPriceRaw:   trade.EntryPrice.Raw(),
		ExitPriceRaw:    trade.ExitPrice.Raw(),
		NetPnL:          trade.NetPnL,
		ExitReason:      trade.ExitReason,
		WorkflowName:    trade.WorkflowName,
		WorkflowVersion: trade.WorkflowVersion,
		ParameterSetID:  trade.ParameterSetID,
	}
	if !trade.EntryTime.IsZero() {
		view.EntryTime = trade.EntryTime.Time().UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	if !trade.ExitTime.IsZero() {
		view.ExitTime = trade.ExitTime.Time().UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	return view
}
