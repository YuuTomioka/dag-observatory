package marketdata

import (
	"fmt"
	"time"

	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	domainmarketdata "dag-observatory/dag-core/internal/domain/marketdata"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

func ToBulkUpsertSessionBarsParams(
	bars []domainmarketdata.SessionBar,
) (query.BulkUpsertSessionBarsParams, error) {
	symbolIDs := make([]int64, 0, len(bars))
	sessionCodes := make([]string, 0, len(bars))
	sessionDates := make([]pgtype.Date, 0, len(bars))
	openTimes := make([]pgtype.Timestamptz, 0, len(bars))
	closeTimes := make([]pgtype.Timestamptz, 0, len(bars))
	opens := make([]int64, 0, len(bars))
	highs := make([]int64, 0, len(bars))
	highTimes := make([]pgtype.Timestamptz, 0, len(bars))
	lows := make([]int64, 0, len(bars))
	lowTimes := make([]pgtype.Timestamptz, 0, len(bars))
	closes := make([]int64, 0, len(bars))
	volumes := make([]int64, 0, len(bars))
	sources := make([]string, 0, len(bars))

	for _, bar := range bars {
		if bar.SymbolID <= 0 {
			return query.BulkUpsertSessionBarsParams{}, fmt.Errorf("%w: session bar symbol_id must be > 0", apprepository.ErrInvalidArgument)
		}
		if bar.SessionCode == "" {
			return query.BulkUpsertSessionBarsParams{}, fmt.Errorf("%w: session bar session_code is required", apprepository.ErrInvalidArgument)
		}
		sessionDate, err := toPgDate(bar.SessionDate)
		if err != nil {
			return query.BulkUpsertSessionBarsParams{}, err
		}
		openTime, err := toPgTimestamptz(bar.Opentime)
		if err != nil {
			return query.BulkUpsertSessionBarsParams{}, err
		}
		closeTime, err := toPgTimestamptz(bar.Closetime)
		if err != nil {
			return query.BulkUpsertSessionBarsParams{}, err
		}
		highTime, err := toPgTimestamptz(bar.Hightime)
		if err != nil {
			return query.BulkUpsertSessionBarsParams{}, err
		}
		lowTime, err := toPgTimestamptz(bar.Lowtime)
		if err != nil {
			return query.BulkUpsertSessionBarsParams{}, err
		}

		symbolIDs = append(symbolIDs, int64(bar.SymbolID))
		sessionCodes = append(sessionCodes, string(bar.SessionCode))
		sessionDates = append(sessionDates, sessionDate)
		openTimes = append(openTimes, openTime)
		closeTimes = append(closeTimes, closeTime)
		opens = append(opens, bar.Open.Raw())
		highs = append(highs, bar.High.Raw())
		highTimes = append(highTimes, highTime)
		lows = append(lows, bar.Low.Raw())
		lowTimes = append(lowTimes, lowTime)
		closes = append(closes, bar.Close.Raw())
		volumes = append(volumes, int64(bar.Volume))
		sources = append(sources, bar.Source)
	}

	return query.BulkUpsertSessionBarsParams{
		SymbolIds:    symbolIDs,
		SessionCodes: sessionCodes,
		SessionDates: sessionDates,
		OpenTimes:    openTimes,
		CloseTimes:   closeTimes,
		Opens:        opens,
		Highs:        highs,
		HighTimes:    highTimes,
		Lows:         lows,
		LowTimes:     lowTimes,
		Closes:       closes,
		Volumes:      volumes,
		Sources:      sources,
	}, nil
}

func ToDeleteSessionBarsBySymbolSessionAndDateRangeParams(
	symbolID domainmarketdata.SymbolID,
	sessionCode domainmarketdata.SessionCode,
	from domainmarketdata.SessionDate,
	to domainmarketdata.SessionDate,
) (query.DeleteSessionBarsBySymbolSessionAndDateRangeParams, error) {
	fromDate, err := toPgDate(from)
	if err != nil {
		return query.DeleteSessionBarsBySymbolSessionAndDateRangeParams{}, err
	}
	toDate, err := toPgDate(to)
	if err != nil {
		return query.DeleteSessionBarsBySymbolSessionAndDateRangeParams{}, err
	}
	return query.DeleteSessionBarsBySymbolSessionAndDateRangeParams{
		SymbolID:    int64(symbolID),
		SessionCode: string(sessionCode),
		FromDate:    fromDate,
		ToDate:      toDate,
	}, nil
}

func ToGetLatestSessionBarBySymbolAndSessionParams(
	symbolID domainmarketdata.SymbolID,
	sessionCode domainmarketdata.SessionCode,
) query.GetLatestSessionBarBySymbolAndSessionParams {
	return query.GetLatestSessionBarBySymbolAndSessionParams{
		SymbolID:    int64(symbolID),
		SessionCode: string(sessionCode),
	}
}

func ToListSessionBarsBySymbolSessionAndDateRangeParams(
	symbolID domainmarketdata.SymbolID,
	sessionCode domainmarketdata.SessionCode,
	from domainmarketdata.SessionDate,
	to domainmarketdata.SessionDate,
) (query.ListSessionBarsBySymbolSessionAndDateRangeParams, error) {
	fromDate, err := toPgDate(from)
	if err != nil {
		return query.ListSessionBarsBySymbolSessionAndDateRangeParams{}, err
	}
	toDate, err := toPgDate(to)
	if err != nil {
		return query.ListSessionBarsBySymbolSessionAndDateRangeParams{}, err
	}
	return query.ListSessionBarsBySymbolSessionAndDateRangeParams{
		SymbolID:    int64(symbolID),
		SessionCode: string(sessionCode),
		FromDate:    fromDate,
		ToDate:      toDate,
	}, nil
}

func SessionBarFromRow(row query.SessionBar) (domainmarketdata.SessionBar, error) {
	sessionDate, err := fromPgDate(row.SessionDate)
	if err != nil {
		return domainmarketdata.SessionBar{}, err
	}
	openTime, err := fromPgTimestamptz(row.OpenTime)
	if err != nil {
		return domainmarketdata.SessionBar{}, err
	}
	closeTime, err := fromPgTimestamptz(row.CloseTime)
	if err != nil {
		return domainmarketdata.SessionBar{}, err
	}
	highTime, err := fromPgTimestamptz(row.HighTime)
	if err != nil {
		return domainmarketdata.SessionBar{}, err
	}
	lowTime, err := fromPgTimestamptz(row.LowTime)
	if err != nil {
		return domainmarketdata.SessionBar{}, err
	}
	return domainmarketdata.SessionBar{
		SymbolID:    domainmarketdata.SymbolID(row.SymbolID),
		SessionCode: domainmarketdata.SessionCode(row.SessionCode),
		SessionDate: sessionDate,
		OHLCV: domainmarketdata.OHLCV{
			Opentime:  openTime,
			Closetime: closeTime,
			Open:      domainmarketdata.NewPriceFromRaw(row.Open),
			High:      domainmarketdata.NewPriceFromRaw(row.High),
			Hightime:  highTime,
			Low:       domainmarketdata.NewPriceFromRaw(row.Low),
			Lowtime:   lowTime,
			Close:     domainmarketdata.NewPriceFromRaw(row.Close),
			Volume:    domainmarketdata.Volume(row.Volume),
		},
		Source: row.Source,
	}, nil
}

func SessionBarsFromRows(rows []query.SessionBar) ([]domainmarketdata.SessionBar, error) {
	out := make([]domainmarketdata.SessionBar, 0, len(rows))
	for _, row := range rows {
		mapped, err := SessionBarFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}
	return out, nil
}

func toPgDate(d domainmarketdata.SessionDate) (pgtype.Date, error) {
	if d.IsZero() {
		return pgtype.Date{}, fmt.Errorf("%w: session date is required", apprepository.ErrInvalidArgument)
	}
	return pgtype.Date{
		Time:  time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC),
		Valid: true,
	}, nil
}

func fromPgDate(v pgtype.Date) (domainmarketdata.SessionDate, error) {
	if !v.Valid {
		return domainmarketdata.SessionDate{}, fmt.Errorf("%w: session date is NULL", apprepository.ErrInvalidArgument)
	}
	return domainmarketdata.NewSessionDate(v.Time.Year(), v.Time.Month(), v.Time.Day())
}
