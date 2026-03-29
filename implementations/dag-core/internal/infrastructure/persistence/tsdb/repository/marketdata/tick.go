package marketdata

import (
	"context"
	"fmt"
	"time"

	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

type Tick struct {
	SymbolID int64
	Time     time.Time
	Bid      int64
	Ask      int64
}

type TickRepository struct {
	client  *tsdb.Client
	queries *query.Queries
}

func NewTickRepository(client *tsdb.Client) *TickRepository {
	var q *query.Queries
	if client != nil && client.Pool != nil {
		q = query.New(client.Pool)
	}
	return &TickRepository{
		client:  client,
		queries: q,
	}
}

func (r *TickRepository) Insert(ctx context.Context, tick Tick) error {
	if err := r.validate(); err != nil {
		return err
	}
	return r.queries.InsertTick(ctx, query.InsertTickParams{
		SymbolID: tick.SymbolID,
		Time:     toPgTimestamptz(tick.Time),
		Bid:      tick.Bid,
		Ask:      tick.Ask,
	})
}

func (r *TickRepository) Upsert(ctx context.Context, tick Tick) error {
	if err := r.validate(); err != nil {
		return err
	}
	return r.queries.UpsertTick(ctx, query.UpsertTickParams{
		SymbolID: tick.SymbolID,
		Time:     toPgTimestamptz(tick.Time),
		Bid:      tick.Bid,
		Ask:      tick.Ask,
	})
}

func (r *TickRepository) BulkUpsert(ctx context.Context, ticks []Tick) error {
	if err := r.validate(); err != nil {
		return err
	}
	if len(ticks) == 0 {
		return nil
	}

	symbolIDs := make([]int64, 0, len(ticks))
	times := make([]pgtype.Timestamptz, 0, len(ticks))
	bids := make([]int64, 0, len(ticks))
	asks := make([]int64, 0, len(ticks))

	for _, tick := range ticks {
		symbolIDs = append(symbolIDs, tick.SymbolID)
		times = append(times, toPgTimestamptz(tick.Time))
		bids = append(bids, tick.Bid)
		asks = append(asks, tick.Ask)
	}

	return r.queries.BulkUpsertTicks(ctx, query.BulkUpsertTicksParams{
		SymbolIds: symbolIDs,
		Times:     times,
		Bids:      bids,
		Asks:      asks,
	})
}

func (r *TickRepository) GetLatestBySymbol(ctx context.Context, symbolID int64) (Tick, error) {
	if err := r.validate(); err != nil {
		return Tick{}, err
	}
	row, err := r.queries.GetLatestTickBySymbol(ctx, symbolID)
	if err != nil {
		return Tick{}, err
	}
	return mapTick(row)
}

func (r *TickRepository) ListBySymbolAndRange(ctx context.Context, symbolID int64, from, to time.Time) ([]Tick, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListTicksBySymbolAndRange(ctx, query.ListTicksBySymbolAndRangeParams{
		SymbolID: symbolID,
		FromTime: toPgTimestamptz(from),
		ToTime:   toPgTimestamptz(to),
	})
	if err != nil {
		return nil, err
	}
	return mapTicks(rows)
}

func (r *TickRepository) ListBySymbolsAndRange(ctx context.Context, symbolIDs []int64, from, to time.Time) ([]Tick, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListTicksBySymbolsAndRange(ctx, query.ListTicksBySymbolsAndRangeParams{
		SymbolIds: symbolIDs,
		FromTime:  toPgTimestamptz(from),
		ToTime:    toPgTimestamptz(to),
	})
	if err != nil {
		return nil, err
	}
	return mapTicks(rows)
}

func (r *TickRepository) ListByRange(ctx context.Context, from, to time.Time) ([]Tick, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListTicksByRange(ctx, query.ListTicksByRangeParams{
		FromTime: toPgTimestamptz(from),
		ToTime:   toPgTimestamptz(to),
	})
	if err != nil {
		return nil, err
	}
	return mapTicks(rows)
}

func (r *TickRepository) validate() error {
	if r == nil || r.client == nil || r.client.Pool == nil || r.queries == nil {
		return fmt.Errorf("tsdb: client not configured")
	}
	return nil
}

func mapTicks(rows []query.Tick) ([]Tick, error) {
	out := make([]Tick, 0, len(rows))
	for _, row := range rows {
		mapped, err := mapTick(row)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}
	return out, nil
}

func mapTick(row query.Tick) (Tick, error) {
	tm, err := fromPgTimestamptz(row.Time)
	if err != nil {
		return Tick{}, err
	}
	return Tick{
		SymbolID: row.SymbolID,
		Time:     tm,
		Bid:      row.Bid,
		Ask:      row.Ask,
	}, nil
}

func toPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}

func fromPgTimestamptz(v pgtype.Timestamptz) (time.Time, error) {
	if !v.Valid {
		return time.Time{}, fmt.Errorf("tsdb: tick time is NULL")
	}
	return v.Time, nil
}
