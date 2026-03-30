package usecase

import (
	"context"
	"fmt"
	"strings"

	"dag-observatory/dag-core/internal/application/marketdata/repository"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

type BackfillMode string

const (
	BackfillModeReplaceRange BackfillMode = "replace_range"
	BackfillModeUpsertOnly   BackfillMode = "upsert_only"
	defaultChunkSizeBars                  = 1000
)

type BackfillTimeframeBarsRequest struct {
	SymbolID      marketdata.SymbolID
	SymbolCode    string
	TimeframeCode marketdata.TimeframeCode
	From          marketdata.UTCTime
	To            marketdata.UTCTime
	Mode          BackfillMode
	ChunkSizeBars int
}

type BackfillTimeframeBarsResult struct {
	SymbolID      marketdata.SymbolID
	TimeframeCode marketdata.TimeframeCode
	From          marketdata.UTCTime
	To            marketdata.UTCTime
	ChunkCount    int
	BarCount      int
}

type BackfillTimeframeBars struct {
	UnitOfWork repository.UnitOfWork
}

func (u *BackfillTimeframeBars) Execute(
	ctx context.Context,
	req BackfillTimeframeBarsRequest,
) (BackfillTimeframeBarsResult, error) {
	if u == nil || u.UnitOfWork == nil {
		return BackfillTimeframeBarsResult{}, repository.ErrNotConfigured
	}
	req, err := normalizeBackfillRequest(req)
	if err != nil {
		return BackfillTimeframeBarsResult{}, err
	}

	symbolID, err := u.resolveSymbolID(ctx, req)
	if err != nil {
		return BackfillTimeframeBarsResult{}, err
	}

	rangeFrom, rangeTo, err := req.TimeframeCode.BackfillRange(req.From, req.To)
	if err != nil {
		return BackfillTimeframeBarsResult{}, fmt.Errorf("%w: %v", repository.ErrInvalidArgument, err)
	}

	chunks := buildBackfillChunks(req.TimeframeCode, rangeFrom, rangeTo, req.ChunkSizeBars)
	totalBars := 0
	for _, chunk := range chunks {
		var chunkBars int
		err := u.UnitOfWork.Do(ctx, func(repos repository.Repositories) error {
			ticks, err := repos.Ticks().ListBySymbolAndRange(ctx, symbolID, chunk.From, chunk.To)
			if err != nil {
				return err
			}
			bars := marketdata.AggregateTimeframeBars(req.TimeframeCode, symbolID, ticks)
			if req.Mode == BackfillModeReplaceRange {
				if err := repos.TimeframeBars().DeleteBySymbolTimeframeAndRange(
					ctx,
					symbolID,
					req.TimeframeCode,
					chunk.From,
					chunk.To,
				); err != nil {
					return err
				}
			}
			if err := repos.TimeframeBars().BulkUpsert(ctx, bars); err != nil {
				return err
			}
			chunkBars = len(bars)
			return nil
		})
		if err != nil {
			return BackfillTimeframeBarsResult{}, err
		}
		totalBars += chunkBars
	}

	return BackfillTimeframeBarsResult{
		SymbolID:      symbolID,
		TimeframeCode: req.TimeframeCode,
		From:          rangeFrom,
		To:            rangeTo,
		ChunkCount:    len(chunks),
		BarCount:      totalBars,
	}, nil
}

type GetLatestTimeframeBarBySymbolAndTimeframe struct {
	UnitOfWork repository.UnitOfWork
}

func (u *GetLatestTimeframeBarBySymbolAndTimeframe) Execute(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode marketdata.TimeframeCode,
) (marketdata.TimeframeBar, error) {
	if u == nil || u.UnitOfWork == nil {
		return marketdata.TimeframeBar{}, repository.ErrNotConfigured
	}
	if symbolID <= 0 {
		return marketdata.TimeframeBar{}, fmt.Errorf("%w: symbol_id must be > 0", repository.ErrInvalidArgument)
	}
	if timeframeCode == "" {
		return marketdata.TimeframeBar{}, fmt.Errorf("%w: timeframe_code is required", repository.ErrInvalidArgument)
	}

	var out marketdata.TimeframeBar
	err := u.UnitOfWork.DoReadOnly(ctx, func(repos repository.Repositories) error {
		item, err := repos.TimeframeBars().GetLatestBySymbolAndTimeframe(ctx, symbolID, timeframeCode)
		if err != nil {
			return err
		}
		out = item
		return nil
	})
	if err != nil {
		return marketdata.TimeframeBar{}, err
	}
	return out, nil
}

type ListTimeframeBarsBySymbolTimeframeAndRange struct {
	UnitOfWork repository.UnitOfWork
}

func (u *ListTimeframeBarsBySymbolTimeframeAndRange) Execute(
	ctx context.Context,
	symbolID marketdata.SymbolID,
	timeframeCode marketdata.TimeframeCode,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
) ([]marketdata.TimeframeBar, error) {
	if u == nil || u.UnitOfWork == nil {
		return nil, repository.ErrNotConfigured
	}
	if symbolID <= 0 {
		return nil, fmt.Errorf("%w: symbol_id must be > 0", repository.ErrInvalidArgument)
	}
	if timeframeCode == "" {
		return nil, fmt.Errorf("%w: timeframe_code is required", repository.ErrInvalidArgument)
	}
	if from.IsZero() || to.IsZero() {
		return nil, fmt.Errorf("%w: from/to are required", repository.ErrInvalidArgument)
	}
	if !from.Before(to) {
		return nil, fmt.Errorf("%w: from must be before to", repository.ErrInvalidArgument)
	}

	var out []marketdata.TimeframeBar
	err := u.UnitOfWork.DoReadOnly(ctx, func(repos repository.Repositories) error {
		items, err := repos.TimeframeBars().ListBySymbolTimeframeAndRange(ctx, symbolID, timeframeCode, from, to)
		if err != nil {
			return err
		}
		out = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

type backfillChunk struct {
	From marketdata.UTCTime
	To   marketdata.UTCTime
}

func normalizeBackfillRequest(req BackfillTimeframeBarsRequest) (BackfillTimeframeBarsRequest, error) {
	if req.TimeframeCode == "" {
		return req, fmt.Errorf("%w: timeframe_code is required", repository.ErrInvalidArgument)
	}
	if _, ok := marketdata.TimeframeDefs[req.TimeframeCode]; !ok {
		return req, fmt.Errorf("%w: timeframe_code %q is not supported", repository.ErrInvalidArgument, req.TimeframeCode)
	}
	if req.From.IsZero() || req.To.IsZero() {
		return req, fmt.Errorf("%w: from/to are required", repository.ErrInvalidArgument)
	}
	if !req.From.Before(req.To) {
		return req, fmt.Errorf("%w: from must be before to", repository.ErrInvalidArgument)
	}
	req.SymbolCode = strings.TrimSpace(strings.ToUpper(req.SymbolCode))
	if req.SymbolID <= 0 && req.SymbolCode == "" {
		return req, fmt.Errorf("%w: symbol_id or symbol_code is required", repository.ErrInvalidArgument)
	}
	switch req.Mode {
	case "", BackfillModeReplaceRange:
		req.Mode = BackfillModeReplaceRange
	case BackfillModeUpsertOnly:
	default:
		return req, fmt.Errorf("%w: mode must be replace_range or upsert_only", repository.ErrInvalidArgument)
	}
	if req.ChunkSizeBars == 0 {
		req.ChunkSizeBars = defaultChunkSizeBars
	}
	if req.ChunkSizeBars <= 0 {
		return req, fmt.Errorf("%w: chunk_size_bars must be > 0", repository.ErrInvalidArgument)
	}
	return req, nil
}

func (u *BackfillTimeframeBars) resolveSymbolID(
	ctx context.Context,
	req BackfillTimeframeBarsRequest,
) (marketdata.SymbolID, error) {
	if req.SymbolID > 0 {
		return req.SymbolID, nil
	}
	var symbolID marketdata.SymbolID
	err := u.UnitOfWork.DoReadOnly(ctx, func(repos repository.Repositories) error {
		symbol, err := repos.Symbols().GetByCode(ctx, req.SymbolCode)
		if err != nil {
			return err
		}
		symbolID = symbol.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return symbolID, nil
}

func buildBackfillChunks(
	timeframeCode marketdata.TimeframeCode,
	from marketdata.UTCTime,
	to marketdata.UTCTime,
	chunkSizeBars int,
) []backfillChunk {
	chunks := make([]backfillChunk, 0)
	cursor := from
	for cursor.Before(to) {
		chunkEnd := cursor
		for i := 0; i < chunkSizeBars && chunkEnd.Before(to); i++ {
			_, next := timeframeCode.Window(chunkEnd)
			chunkEnd = next
		}
		if chunkEnd.After(to) {
			chunkEnd = to
		}
		chunks = append(chunks, backfillChunk{From: cursor, To: chunkEnd})
		cursor = chunkEnd
	}
	return chunks
}
