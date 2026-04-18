package marketdata

import (
	"fmt"

	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	domainmarketdata "dag-observatory/dag-core/internal/domain/marketdata"
	domainmarketphase "dag-observatory/dag-core/internal/domain/marketphase"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

func ToBulkUpsertPhaseBarsParams(
	bars []domainmarketphase.PhaseBar,
) (query.BulkUpsertPhaseBarsParams, error) {
	symbolIDs := make([]int64, 0, len(bars))
	phaseIDs := make([]string, 0, len(bars))
	markets := make([]string, 0, len(bars))
	timezones := make([]string, 0, len(bars))
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
			return query.BulkUpsertPhaseBarsParams{}, fmt.Errorf("%w: phase bar symbol_id must be > 0", apprepository.ErrInvalidArgument)
		}
		if bar.PhaseID == "" {
			return query.BulkUpsertPhaseBarsParams{}, fmt.Errorf("%w: phase bar phase_id is required", apprepository.ErrInvalidArgument)
		}
		if bar.Market == "" {
			return query.BulkUpsertPhaseBarsParams{}, fmt.Errorf("%w: phase bar market is required", apprepository.ErrInvalidArgument)
		}
		if bar.Timezone == "" {
			return query.BulkUpsertPhaseBarsParams{}, fmt.Errorf("%w: phase bar timezone is required", apprepository.ErrInvalidArgument)
		}
		openTime, err := toPgTimestamptz(bar.Opentime)
		if err != nil {
			return query.BulkUpsertPhaseBarsParams{}, err
		}
		closeTime, err := toPgTimestamptz(bar.Closetime)
		if err != nil {
			return query.BulkUpsertPhaseBarsParams{}, err
		}
		highTime, err := toPgTimestamptz(bar.Hightime)
		if err != nil {
			return query.BulkUpsertPhaseBarsParams{}, err
		}
		lowTime, err := toPgTimestamptz(bar.Lowtime)
		if err != nil {
			return query.BulkUpsertPhaseBarsParams{}, err
		}

		symbolIDs = append(symbolIDs, int64(bar.SymbolID))
		phaseIDs = append(phaseIDs, string(bar.PhaseID))
		markets = append(markets, bar.Market)
		timezones = append(timezones, bar.Timezone)
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

	return query.BulkUpsertPhaseBarsParams{
		SymbolIds:  symbolIDs,
		PhaseIds:   phaseIDs,
		Markets:    markets,
		Timezones:  timezones,
		OpenTimes:  openTimes,
		CloseTimes: closeTimes,
		Opens:      opens,
		Highs:      highs,
		HighTimes:  highTimes,
		Lows:       lows,
		LowTimes:   lowTimes,
		Closes:     closes,
		Volumes:    volumes,
		Sources:    sources,
	}, nil
}

func ToDeletePhaseBarsBySymbolPhaseAndRangeParams(
	symbolID domainmarketdata.SymbolID,
	phaseID domainmarketphase.PhaseID,
	from domainmarketdata.UTCTime,
	to domainmarketdata.UTCTime,
) (query.DeletePhaseBarsBySymbolPhaseAndRangeParams, error) {
	fromTime, err := toPgTimestamptz(from)
	if err != nil {
		return query.DeletePhaseBarsBySymbolPhaseAndRangeParams{}, err
	}
	toTime, err := toPgTimestamptz(to)
	if err != nil {
		return query.DeletePhaseBarsBySymbolPhaseAndRangeParams{}, err
	}
	return query.DeletePhaseBarsBySymbolPhaseAndRangeParams{
		SymbolID: int64(symbolID),
		PhaseID:  string(phaseID),
		FromTime: fromTime,
		ToTime:   toTime,
	}, nil
}

func ToGetLatestPhaseBarBySymbolAndPhaseParams(
	symbolID domainmarketdata.SymbolID,
	phaseID domainmarketphase.PhaseID,
) query.GetLatestPhaseBarBySymbolAndPhaseParams {
	return query.GetLatestPhaseBarBySymbolAndPhaseParams{
		SymbolID: int64(symbolID),
		PhaseID:  string(phaseID),
	}
}

func ToListPhaseBarsBySymbolPhaseAndRangeParams(
	symbolID domainmarketdata.SymbolID,
	phaseID domainmarketphase.PhaseID,
	from domainmarketdata.UTCTime,
	to domainmarketdata.UTCTime,
) (query.ListPhaseBarsBySymbolPhaseAndRangeParams, error) {
	fromTime, err := toPgTimestamptz(from)
	if err != nil {
		return query.ListPhaseBarsBySymbolPhaseAndRangeParams{}, err
	}
	toTime, err := toPgTimestamptz(to)
	if err != nil {
		return query.ListPhaseBarsBySymbolPhaseAndRangeParams{}, err
	}
	return query.ListPhaseBarsBySymbolPhaseAndRangeParams{
		SymbolID: int64(symbolID),
		PhaseID:  string(phaseID),
		FromTime: fromTime,
		ToTime:   toTime,
	}, nil
}

func PhaseBarFromRow(row query.PhaseBar) (domainmarketphase.PhaseBar, error) {
	openTime, err := fromPgTimestamptz(row.OpenTime)
	if err != nil {
		return domainmarketphase.PhaseBar{}, err
	}
	closeTime, err := fromPgTimestamptz(row.CloseTime)
	if err != nil {
		return domainmarketphase.PhaseBar{}, err
	}
	highTime, err := fromPgTimestamptz(row.HighTime)
	if err != nil {
		return domainmarketphase.PhaseBar{}, err
	}
	lowTime, err := fromPgTimestamptz(row.LowTime)
	if err != nil {
		return domainmarketphase.PhaseBar{}, err
	}
	return domainmarketphase.PhaseBar{
		PhaseID:  domainmarketphase.PhaseID(row.PhaseID),
		Market:   row.Market,
		Timezone: row.Timezone,
		SymbolID: domainmarketdata.SymbolID(row.SymbolID),
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

func PhaseBarsFromRows(rows []query.PhaseBar) ([]domainmarketphase.PhaseBar, error) {
	out := make([]domainmarketphase.PhaseBar, 0, len(rows))
	for _, row := range rows {
		mapped, err := PhaseBarFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}
	return out, nil
}
