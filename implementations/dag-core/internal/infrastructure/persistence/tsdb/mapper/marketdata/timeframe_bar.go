package marketdata

import (
	"fmt"

	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	domainmarketdata "dag-observatory/dag-core/internal/domain/marketdata"
	domainohlc "dag-observatory/dag-core/internal/domain/marketdata/ohlc"
	domaintimeframe "dag-observatory/dag-core/internal/domain/marketdata/ohlc/timeframe"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

func ToBulkUpsertTimeframeBarsParams(
	bars []domaintimeframe.TimeframeBar,
) (query.BulkUpsertTimeframeBarsParams, error) {
	symbolIDs := make([]int64, 0, len(bars))
	timeframeCodes := make([]string, 0, len(bars))
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
		openTime, err := toPgTimestamptz(bar.Opentime)
		if err != nil {
			return query.BulkUpsertTimeframeBarsParams{}, err
		}
		closeTime, err := toPgTimestamptz(bar.Closetime)
		if err != nil {
			return query.BulkUpsertTimeframeBarsParams{}, err
		}
		highTime, err := toPgTimestamptz(bar.Hightime)
		if err != nil {
			return query.BulkUpsertTimeframeBarsParams{}, err
		}
		lowTime, err := toPgTimestamptz(bar.Lowtime)
		if err != nil {
			return query.BulkUpsertTimeframeBarsParams{}, err
		}
		if bar.SymbolID <= 0 {
			return query.BulkUpsertTimeframeBarsParams{}, fmt.Errorf("%w: timeframe bar symbol_id must be > 0", apprepository.ErrInvalidArgument)
		}
		if bar.TimeframeCode == "" {
			return query.BulkUpsertTimeframeBarsParams{}, fmt.Errorf("%w: timeframe bar timeframe_code is required", apprepository.ErrInvalidArgument)
		}
		symbolIDs = append(symbolIDs, int64(bar.SymbolID))
		timeframeCodes = append(timeframeCodes, string(bar.TimeframeCode))
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
	return query.BulkUpsertTimeframeBarsParams{
		SymbolIds:      symbolIDs,
		TimeframeCodes: timeframeCodes,
		OpenTimes:      openTimes,
		CloseTimes:     closeTimes,
		Opens:          opens,
		Highs:          highs,
		HighTimes:      highTimes,
		Lows:           lows,
		LowTimes:       lowTimes,
		Closes:         closes,
		Volumes:        volumes,
		Sources:        sources,
	}, nil
}

func ToDeleteTimeframeBarsBySymbolTimeframeAndRangeParams(
	symbolID domainmarketdata.SymbolID,
	timeframeCode domaintimeframe.TimeframeCode,
	from domainmarketdata.UTCTime,
	to domainmarketdata.UTCTime,
) (query.DeleteTimeframeBarsBySymbolTimeframeAndRangeParams, error) {
	fromTime, err := toPgTimestamptz(from)
	if err != nil {
		return query.DeleteTimeframeBarsBySymbolTimeframeAndRangeParams{}, err
	}
	toTime, err := toPgTimestamptz(to)
	if err != nil {
		return query.DeleteTimeframeBarsBySymbolTimeframeAndRangeParams{}, err
	}
	return query.DeleteTimeframeBarsBySymbolTimeframeAndRangeParams{
		SymbolID:      int64(symbolID),
		TimeframeCode: string(timeframeCode),
		FromTime:      fromTime,
		ToTime:        toTime,
	}, nil
}

func ToGetLatestTimeframeBarBySymbolAndTimeframeParams(
	symbolID domainmarketdata.SymbolID,
	timeframeCode domaintimeframe.TimeframeCode,
) query.GetLatestTimeframeBarBySymbolAndTimeframeParams {
	return query.GetLatestTimeframeBarBySymbolAndTimeframeParams{
		SymbolID:      int64(symbolID),
		TimeframeCode: string(timeframeCode),
	}
}

func ToListTimeframeBarsBySymbolTimeframeAndRangeParams(
	symbolID domainmarketdata.SymbolID,
	timeframeCode domaintimeframe.TimeframeCode,
	from domainmarketdata.UTCTime,
	to domainmarketdata.UTCTime,
) (query.ListTimeframeBarsBySymbolTimeframeAndRangeParams, error) {
	fromTime, err := toPgTimestamptz(from)
	if err != nil {
		return query.ListTimeframeBarsBySymbolTimeframeAndRangeParams{}, err
	}
	toTime, err := toPgTimestamptz(to)
	if err != nil {
		return query.ListTimeframeBarsBySymbolTimeframeAndRangeParams{}, err
	}
	return query.ListTimeframeBarsBySymbolTimeframeAndRangeParams{
		SymbolID:      int64(symbolID),
		TimeframeCode: string(timeframeCode),
		FromTime:      fromTime,
		ToTime:        toTime,
	}, nil
}

func TimeframeBarFromRow(row query.TimeframeBar) (domaintimeframe.TimeframeBar, error) {
	openTime, err := fromPgTimestamptz(row.OpenTime)
	if err != nil {
		return domaintimeframe.TimeframeBar{}, err
	}
	closeTime, err := fromPgTimestamptz(row.CloseTime)
	if err != nil {
		return domaintimeframe.TimeframeBar{}, err
	}
	highTime, err := fromPgTimestamptz(row.HighTime)
	if err != nil {
		return domaintimeframe.TimeframeBar{}, err
	}
	lowTime, err := fromPgTimestamptz(row.LowTime)
	if err != nil {
		return domaintimeframe.TimeframeBar{}, err
	}
	return domaintimeframe.TimeframeBar{
		SymbolID:      domainmarketdata.SymbolID(row.SymbolID),
		TimeframeCode: domaintimeframe.TimeframeCode(row.TimeframeCode),
		OHLCV: domainohlc.OHLCV{
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

func TimeframeBarsFromRows(rows []query.TimeframeBar) ([]domaintimeframe.TimeframeBar, error) {
	out := make([]domaintimeframe.TimeframeBar, 0, len(rows))
	for _, row := range rows {
		mapped, err := TimeframeBarFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}
	return out, nil
}
