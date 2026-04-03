package factory

import (
	"context"
	"fmt"
	"strings"

	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	marketdatarepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

func normalizeResolvedMarketdataInput(av artifact.View) (resolvedMarketdataInput, error) {
	timeframeCode := marketdata.TimeframeCode(artifact.MustGet(av, dagruntimeusecase.InputKeyMarketdataTimeframeCode))
	from := artifact.MustGet(av, dagruntimeusecase.InputKeyMarketdataFrom)
	to := artifact.MustGet(av, dagruntimeusecase.InputKeyMarketdataTo)

	symbolCode, _ := artifact.Get(av, dagruntimeusecase.InputKeyMarketdataSymbolCode)
	symbolCode = strings.TrimSpace(strings.ToUpper(symbolCode))

	var symbolID marketdata.SymbolID
	if rawID, ok := artifact.Get(av, dagruntimeusecase.InputKeyMarketdataSymbolID); ok && rawID > 0 {
		symbolID = marketdata.SymbolID(rawID)
	}

	if _, ok := marketdata.TimeframeDefs[timeframeCode]; !ok {
		return resolvedMarketdataInput{}, fmt.Errorf("%w: timeframe_code %q is not supported", marketdatarepository.ErrInvalidArgument, timeframeCode)
	}
	if from.IsZero() || to.IsZero() {
		return resolvedMarketdataInput{}, fmt.Errorf("%w: from/to are required", marketdatarepository.ErrInvalidArgument)
	}
	if !from.Before(to) {
		return resolvedMarketdataInput{}, fmt.Errorf("%w: from must be before to", marketdatarepository.ErrInvalidArgument)
	}
	if symbolID <= 0 && symbolCode == "" {
		return resolvedMarketdataInput{}, fmt.Errorf("%w: symbol_id or symbol_code is required", marketdatarepository.ErrInvalidArgument)
	}

	rangeFrom, rangeTo, err := timeframeCode.BackfillRange(from, to)
	if err != nil {
		return resolvedMarketdataInput{}, fmt.Errorf("%w: %v", marketdatarepository.ErrInvalidArgument, err)
	}
	return resolvedMarketdataInput{
		SymbolID:      symbolID,
		SymbolCode:    symbolCode,
		TimeframeCode: timeframeCode,
		From:          rangeFrom,
		To:            rangeTo,
	}, nil
}

func resolveMarketdataSymbolID(
	ctx context.Context,
	uow marketdatarepository.UnitOfWork,
	symbolID marketdata.SymbolID,
	symbolCode string,
) (marketdata.SymbolID, error) {
	if symbolID > 0 {
		return symbolID, nil
	}
	var resolved marketdata.SymbolID
	err := uow.DoReadOnly(ctx, func(repos marketdatarepository.Repositories) error {
		symbol, err := repos.Symbols().GetByCode(ctx, symbolCode)
		if err != nil {
			return err
		}
		resolved = symbol.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return resolved, nil
}

func buildBackfillPlanChunks(
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

func timeframeBarBackfillResultKey(nodeID string) artifact.Key[marketdatausecase.BackfillTimeframeBarsResult] {
	return artifact.Key[marketdatausecase.BackfillTimeframeBarsResult]{
		Name:     fmt.Sprintf("%s.backfill_result", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.marketdata.backfill.%s.v1", nodeID),
	}
}

func resolvedMarketdataInputKey(nodeID string) artifact.Key[resolvedMarketdataInput] {
	return artifact.Key[resolvedMarketdataInput]{
		Name:     fmt.Sprintf("%s.resolved_marketdata", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.marketdata.backfill.%s.resolved.v1", nodeID),
	}
}

func backfillPlanKey(nodeID string) artifact.Key[timeframeBarBackfillPlan] {
	return artifact.Key[timeframeBarBackfillPlan]{
		Name:     fmt.Sprintf("%s.backfill_plan", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.marketdata.backfill.%s.plan.v1", nodeID),
	}
}

func loadedBackfillPlanKey(nodeID string) artifact.Key[timeframeBarLoadedPlan] {
	return artifact.Key[timeframeBarLoadedPlan]{
		Name:     fmt.Sprintf("%s.loaded_backfill_plan", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.marketdata.backfill.%s.loaded.v1", nodeID),
	}
}

func aggregatedBackfillPlanKey(nodeID string) artifact.Key[timeframeBarAggregatedPlan] {
	return artifact.Key[timeframeBarAggregatedPlan]{
		Name:     fmt.Sprintf("%s.aggregated_backfill_plan", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.marketdata.backfill.%s.aggregated.v1", nodeID),
	}
}
