package marketdata

import (
	"fmt"

	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	domainmarketdata "dag-observatory/dag-core/internal/domain/marketdata"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

func ToInsertTickParams(tick domainmarketdata.Tick) (query.InsertTickParams, error) {
	timeValue, err := toPgTimestamptz(tick.Time)
	if err != nil {
		return query.InsertTickParams{}, err
	}
	return query.InsertTickParams{
		SymbolID: int64(tick.SymbolID),
		Time:     timeValue,
		Bid:      tick.Bid.Raw(),
		Ask:      tick.Ask.Raw(),
	}, nil
}

func ToUpsertTickParams(tick domainmarketdata.Tick) (query.UpsertTickParams, error) {
	timeValue, err := toPgTimestamptz(tick.Time)
	if err != nil {
		return query.UpsertTickParams{}, err
	}
	return query.UpsertTickParams{
		SymbolID: int64(tick.SymbolID),
		Time:     timeValue,
		Bid:      tick.Bid.Raw(),
		Ask:      tick.Ask.Raw(),
	}, nil
}

func ToBulkUpsertTickParams(ticks []domainmarketdata.Tick) (query.BulkUpsertTicksParams, error) {
	symbolIDs := make([]int64, 0, len(ticks))
	times := make([]pgtype.Timestamptz, 0, len(ticks))
	bids := make([]int64, 0, len(ticks))
	asks := make([]int64, 0, len(ticks))

	for _, tick := range ticks {
		timeValue, err := toPgTimestamptz(tick.Time)
		if err != nil {
			return query.BulkUpsertTicksParams{}, err
		}
		symbolIDs = append(symbolIDs, int64(tick.SymbolID))
		times = append(times, timeValue)
		bids = append(bids, tick.Bid.Raw())
		asks = append(asks, tick.Ask.Raw())
	}

	return query.BulkUpsertTicksParams{
		SymbolIds: symbolIDs,
		Times:     times,
		Bids:      bids,
		Asks:      asks,
	}, nil
}

func ToListTicksBySymbolAndRangeParams(
	symbolID domainmarketdata.SymbolID,
	from domainmarketdata.UTCTime,
	to domainmarketdata.UTCTime,
) (query.ListTicksBySymbolAndRangeParams, error) {
	fromTime, err := toPgTimestamptz(from)
	if err != nil {
		return query.ListTicksBySymbolAndRangeParams{}, err
	}
	toTime, err := toPgTimestamptz(to)
	if err != nil {
		return query.ListTicksBySymbolAndRangeParams{}, err
	}
	return query.ListTicksBySymbolAndRangeParams{
		SymbolID: int64(symbolID),
		FromTime: fromTime,
		ToTime:   toTime,
	}, nil
}

func ToListTicksBySymbolsAndRangeParams(
	symbolIDs []domainmarketdata.SymbolID,
	from domainmarketdata.UTCTime,
	to domainmarketdata.UTCTime,
) (query.ListTicksBySymbolsAndRangeParams, error) {
	fromTime, err := toPgTimestamptz(from)
	if err != nil {
		return query.ListTicksBySymbolsAndRangeParams{}, err
	}
	toTime, err := toPgTimestamptz(to)
	if err != nil {
		return query.ListTicksBySymbolsAndRangeParams{}, err
	}

	ids := make([]int64, 0, len(symbolIDs))
	for _, symbolID := range symbolIDs {
		ids = append(ids, int64(symbolID))
	}

	return query.ListTicksBySymbolsAndRangeParams{
		SymbolIds: ids,
		FromTime:  fromTime,
		ToTime:    toTime,
	}, nil
}

func ToListTicksByRangeParams(
	from domainmarketdata.UTCTime,
	to domainmarketdata.UTCTime,
) (query.ListTicksByRangeParams, error) {
	fromTime, err := toPgTimestamptz(from)
	if err != nil {
		return query.ListTicksByRangeParams{}, err
	}
	toTime, err := toPgTimestamptz(to)
	if err != nil {
		return query.ListTicksByRangeParams{}, err
	}
	return query.ListTicksByRangeParams{
		FromTime: fromTime,
		ToTime:   toTime,
	}, nil
}

func TickFromRow(row query.Tick) (domainmarketdata.Tick, error) {
	timeValue, err := fromPgTimestamptz(row.Time)
	if err != nil {
		return domainmarketdata.Tick{}, err
	}
	return domainmarketdata.Tick{
		SymbolID: domainmarketdata.SymbolID(row.SymbolID),
		Time:     timeValue,
		Bid:      domainmarketdata.NewPriceFromRaw(row.Bid),
		Ask:      domainmarketdata.NewPriceFromRaw(row.Ask),
	}, nil
}

func TicksFromRows(rows []query.Tick) ([]domainmarketdata.Tick, error) {
	out := make([]domainmarketdata.Tick, 0, len(rows))
	for _, row := range rows {
		mapped, err := TickFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}
	return out, nil
}

func toPgTimestamptz(t domainmarketdata.UTCTime) (pgtype.Timestamptz, error) {
	if t.IsZero() {
		return pgtype.Timestamptz{}, fmt.Errorf("%w: tick time must not be zero", apprepository.ErrInvalidArgument)
	}
	return pgtype.Timestamptz{
		Time:  t.Time(),
		Valid: true,
	}, nil
}

func fromPgTimestamptz(v pgtype.Timestamptz) (domainmarketdata.UTCTime, error) {
	if !v.Valid {
		return domainmarketdata.UTCTime{}, fmt.Errorf("%w: tick time is NULL", apprepository.ErrInvalidArgument)
	}
	return domainmarketdata.NewUTCTime(v.Time), nil
}
