package marketdata

import (
	"fmt"

	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	domainmarketdata "dag-observatory/dag-core/internal/domain/marketdata"
	query "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/query/gen"
)

func ToCreateSymbolParams(input apprepository.CreateSymbolInput) (query.CreateSymbolParams, error) {
	if input.PriceScale > 18 {
		return query.CreateSymbolParams{}, fmt.Errorf("%w: price_scale must be <= 18", apprepository.ErrInvalidArgument)
	}
	if input.TickSizeRaw <= 0 {
		return query.CreateSymbolParams{}, fmt.Errorf("%w: tick_size_raw must be > 0", apprepository.ErrInvalidArgument)
	}
	if input.PipSizeRaw <= 0 {
		return query.CreateSymbolParams{}, fmt.Errorf("%w: pip_size_raw must be > 0", apprepository.ErrInvalidArgument)
	}

	return query.CreateSymbolParams{
		Code:        input.Code,
		Base:        input.Base,
		Quote:       input.Quote,
		PriceScale:  int16(input.PriceScale),
		TickSizeRaw: input.TickSizeRaw,
		PipSizeRaw:  input.PipSizeRaw,
	}, nil
}

func SymbolFromRow(row query.Symbol) (domainmarketdata.Symbol, error) {
	if row.PriceScale < 0 || row.PriceScale > 18 {
		return domainmarketdata.Symbol{}, fmt.Errorf(
			"%w: invalid price_scale=%d from db",
			apprepository.ErrInvalidArgument,
			row.PriceScale,
		)
	}
	if row.TickSizeRaw <= 0 {
		return domainmarketdata.Symbol{}, fmt.Errorf(
			"%w: invalid tick_size_raw=%d from db",
			apprepository.ErrInvalidArgument,
			row.TickSizeRaw,
		)
	}
	if row.PipSizeRaw <= 0 {
		return domainmarketdata.Symbol{}, fmt.Errorf(
			"%w: invalid pip_size_raw=%d from db",
			apprepository.ErrInvalidArgument,
			row.PipSizeRaw,
		)
	}

	return domainmarketdata.Symbol{
		ID:          domainmarketdata.SymbolID(row.ID),
		Code:        row.Code,
		Base:        row.Base,
		Quote:       row.Quote,
		PriceScale:  uint8(row.PriceScale),
		TickSizeRaw: row.TickSizeRaw,
		PipSizeRaw:  row.PipSizeRaw,
	}, nil
}

func SymbolsFromRows(rows []query.Symbol) ([]domainmarketdata.Symbol, error) {
	out := make([]domainmarketdata.Symbol, 0, len(rows))
	for _, row := range rows {
		mapped, err := SymbolFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}
	return out, nil
}
