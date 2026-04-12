package dagruntime

import (
	"context"
	"errors"
	"fmt"
	"time"

	apprepository "dag-observatory/dag-core/internal/application/dagruntime/repository"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type BacktestSummaryRepository struct {
	queries query.Querier
}

func NewBacktestSummaryRepository(client *tsdb.Client) *BacktestSummaryRepository {
	var q query.Querier
	if client != nil && client.Pool != nil {
		q = query.New(client.Pool)
	}
	return NewBacktestSummaryRepositoryWithQuerier(q)
}

func NewBacktestSummaryRepositoryWithQuerier(queries query.Querier) *BacktestSummaryRepository {
	return &BacktestSummaryRepository{queries: queries}
}

func (r *BacktestSummaryRepository) Upsert(ctx context.Context, item apprepository.BacktestRunSummary) error {
	if err := r.validate(); err != nil {
		return err
	}
	return r.queries.UpsertBacktestRunSummary(ctx, query.UpsertBacktestRunSummaryParams{
		RunID:             item.RunID,
		PartitionKey:      item.Partition,
		TradeCount:        item.TradeCount,
		WinRate:           item.WinRate,
		TotalNetPnl:       item.TotalNetPnL,
		MaxDrawdown:       item.MaxDrawdown,
		EquityPointCount:  item.EquityPointCount,
		HasStrategyFields: item.HasStrategyFields,
	})
}

func (r *BacktestSummaryRepository) GetByRunID(ctx context.Context, runID string) (apprepository.BacktestRunSummary, error) {
	if err := r.validate(); err != nil {
		return apprepository.BacktestRunSummary{}, err
	}
	row, err := r.queries.GetBacktestRunSummaryByRunID(ctx, runID)
	if err != nil {
		return apprepository.BacktestRunSummary{}, mapError(err)
	}
	createdAt, err := fromPgTime(row.CreatedAt)
	if err != nil {
		return apprepository.BacktestRunSummary{}, err
	}
	updatedAt, err := fromPgTime(row.UpdatedAt)
	if err != nil {
		return apprepository.BacktestRunSummary{}, err
	}
	return apprepository.BacktestRunSummary{
		RunID:             row.RunID,
		Partition:         row.PartitionKey,
		TradeCount:        row.TradeCount,
		WinRate:           row.WinRate,
		TotalNetPnL:       row.TotalNetPnl,
		MaxDrawdown:       row.MaxDrawdown,
		EquityPointCount:  row.EquityPointCount,
		HasStrategyFields: row.HasStrategyFields,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}, nil
}

func (r *BacktestSummaryRepository) validate() error {
	if r == nil || r.queries == nil {
		return apprepository.ErrNotConfigured
	}
	return nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %v", apprepository.ErrNotFound, err)
	}
	return err
}

func fromPgTime(v pgtype.Timestamptz) (time.Time, error) {
	if !v.Valid {
		return time.Time{}, fmt.Errorf("dagruntime backtest summary repository: timestamp is NULL")
	}
	return v.Time, nil
}

var _ apprepository.BacktestSummaryReader = (*BacktestSummaryRepository)(nil)
var _ apprepository.BacktestSummaryWriter = (*BacktestSummaryRepository)(nil)
