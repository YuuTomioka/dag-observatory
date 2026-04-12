package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	apprepository "dag-observatory/dag-core/internal/application/dagruntime/repository"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type fakeBacktestSummaryRepo struct {
	item        apprepository.BacktestRunSummary
	getErr      error
	upsertCalls []apprepository.BacktestRunSummary
}

func (f *fakeBacktestSummaryRepo) GetByRunID(ctx context.Context, runID string) (apprepository.BacktestRunSummary, error) {
	_ = ctx
	_ = runID
	if f.getErr != nil {
		return apprepository.BacktestRunSummary{}, f.getErr
	}
	return f.item, nil
}

func (f *fakeBacktestSummaryRepo) Upsert(ctx context.Context, item apprepository.BacktestRunSummary) error {
	_ = ctx
	f.upsertCalls = append(f.upsertCalls, item)
	return nil
}

func TestGetRunBacktestSummaryExecute(t *testing.T) {
	now := time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)
	reader := &fakeRunReader{
		runs: []port.RunRecord{
			{
				RunID:     "run-1",
				Partition: state.Partition("p1"),
				EventTime: now,
				StartedAt: now,
				EndedAt:   now.Add(time.Second),
				Duration:  time.Second,
			},
		},
		steps: map[string][]events.NodeExecutionEvent{
			"run-1": {
				{
					SequenceNo: 1,
					StateDiff: []events.StateDiffField{
						{Field: "strategy_summary.trade_count", After: 2.0},
						{Field: "strategy_summary.win_rate", After: 0.5},
						{Field: "strategy_summary.total_net_pnl", After: 8.2},
						{Field: "strategy_summary.max_drawdown", After: 1.7},
						{Field: "equity_series.count", After: 10.0},
					},
				},
			},
		},
	}

	repo := &fakeBacktestSummaryRepo{
		getErr: errors.New("db temporarily unavailable"),
	}
	uc := &GetRunBacktestSummary{
		Reader: reader,
		Writer: repo,
	}
	summary, ok, err := uc.Execute(context.Background(), GetRunBacktestSummaryRequest{RunID: "run-1"})
	if err != nil {
		t.Fatalf("get run backtest summary failed: %v", err)
	}
	if !ok {
		t.Fatal("expected summary to exist")
	}
	if summary.TradeCount != 2 || summary.TotalNetPnL != 8.2 || summary.MaxDrawdown != 1.7 {
		t.Fatalf("unexpected summary values: %#v", summary)
	}
	if summary.EquityPointCount != 10 {
		t.Fatalf("expected equity_point_count=10, got %d", summary.EquityPointCount)
	}
	if len(repo.upsertCalls) != 1 {
		t.Fatalf("expected one persistence upsert call, got %d", len(repo.upsertCalls))
	}
}

func TestGetRunBacktestSummaryExecuteUsesRepositoryProjection(t *testing.T) {
	repo := &fakeBacktestSummaryRepo{
		item: apprepository.BacktestRunSummary{
			RunID:             "run-from-db",
			Partition:         "p-db",
			TradeCount:        9,
			WinRate:           0.66,
			TotalNetPnL:       101.5,
			MaxDrawdown:       12.3,
			EquityPointCount:  200,
			HasStrategyFields: true,
		},
	}
	uc := &GetRunBacktestSummary{
		SummaryRepo: repo,
	}
	summary, ok, err := uc.Execute(context.Background(), GetRunBacktestSummaryRequest{RunID: "run-from-db"})
	if err != nil {
		t.Fatalf("get run backtest summary from repository failed: %v", err)
	}
	if !ok {
		t.Fatal("expected summary to exist")
	}
	if summary.RunID != "run-from-db" || summary.Partition != "p-db" {
		t.Fatalf("unexpected summary identity: %#v", summary)
	}
	if summary.TradeCount != 9 || summary.TotalNetPnL != 101.5 {
		t.Fatalf("unexpected summary values: %#v", summary)
	}
	if len(repo.upsertCalls) != 0 {
		t.Fatalf("expected no upsert when summary is already persisted")
	}
}

func TestGetRunBacktestSummaryExecuteReturnsNotFoundWhenRepoMissAndRunMissing(t *testing.T) {
	uc := &GetRunBacktestSummary{
		SummaryRepo: &fakeBacktestSummaryRepo{getErr: apprepository.ErrNotFound},
		Reader:      &fakeRunReader{},
	}
	_, ok, err := uc.Execute(context.Background(), GetRunBacktestSummaryRequest{RunID: "missing-run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected not found")
	}
}
