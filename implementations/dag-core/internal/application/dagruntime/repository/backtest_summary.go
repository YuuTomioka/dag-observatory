package repository

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("dagruntime backtest summary repository: not found")
	ErrNotConfigured = errors.New("dagruntime backtest summary repository: not configured")
)

type BacktestRunSummary struct {
	RunID             string
	Partition         string
	TradeCount        int64
	WinRate           float64
	TotalNetPnL       float64
	MaxDrawdown       float64
	EquityPointCount  int64
	HasStrategyFields bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type BacktestSummaryReader interface {
	GetByRunID(ctx context.Context, runID string) (BacktestRunSummary, error)
}

type BacktestSummaryWriter interface {
	Upsert(ctx context.Context, item BacktestRunSummary) error
}
