package factory

import (
	"context"

	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	marketdatarepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/domain/marketdata/ohlc/timeframe"
)

type timeframeBarBackfillNode struct {
	id            string
	usecase       *marketdatausecase.BackfillTimeframeBars
	mode          marketdatausecase.BackfillMode
	chunkSizeBars int
}

func (n *timeframeBarBackfillNode) Name() string {
	return "dagruntime.marketdata_timeframe_bar_backfill." + n.id
}

func (n *timeframeBarBackfillNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		dagruntimeusecase.InputKeyMarketdataTimeframeCode,
		dagruntimeusecase.InputKeyMarketdataFrom,
		dagruntimeusecase.InputKeyMarketdataTo,
	}
}

func (n *timeframeBarBackfillNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{timeframeBarBackfillResultKey(n.id)}
}

func (n *timeframeBarBackfillNode) Reads() []state.AnyKey  { return nil }
func (n *timeframeBarBackfillNode) Writes() []state.AnyKey { return nil }
func (n *timeframeBarBackfillNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: false}
}

func (n *timeframeBarBackfillNode) Run(
	ctx context.Context,
	av artifact.View,
	aw artifact.Writer,
	txn state.Txn,
) error {
	_ = txn
	req := marketdatausecase.BackfillTimeframeBarsRequest{
		TimeframeCode: timeframe.TimeframeCode(artifact.MustGet(av, dagruntimeusecase.InputKeyMarketdataTimeframeCode)),
		From:          artifact.MustGet(av, dagruntimeusecase.InputKeyMarketdataFrom),
		To:            artifact.MustGet(av, dagruntimeusecase.InputKeyMarketdataTo),
		Mode:          n.mode,
		ChunkSizeBars: n.chunkSizeBars,
	}
	if symbolID, ok := artifact.Get(av, dagruntimeusecase.InputKeyMarketdataSymbolID); ok && symbolID > 0 {
		req.SymbolID = marketdata.SymbolID(symbolID)
	}
	if symbolCode, ok := artifact.Get(av, dagruntimeusecase.InputKeyMarketdataSymbolCode); ok {
		req.SymbolCode = symbolCode
	}
	result, err := n.usecase.Execute(ctx, req)
	if err != nil {
		return err
	}
	artifact.Set(aw, timeframeBarBackfillResultKey(n.id), result)
	return nil
}

type resolveSymbolNode struct {
	id  string
	uow marketdatarepository.UnitOfWork
}

func (n *resolveSymbolNode) Name() string { return "dagruntime.marketdata_resolve_symbol." + n.id }
func (n *resolveSymbolNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		dagruntimeusecase.InputKeyMarketdataTimeframeCode,
		dagruntimeusecase.InputKeyMarketdataFrom,
		dagruntimeusecase.InputKeyMarketdataTo,
	}
}
func (n *resolveSymbolNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{resolvedMarketdataInputKey(n.id)}
}
func (n *resolveSymbolNode) Reads() []state.AnyKey  { return nil }
func (n *resolveSymbolNode) Writes() []state.AnyKey { return nil }
func (n *resolveSymbolNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: false}
}

func (n *resolveSymbolNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = txn
	input, err := normalizeResolvedMarketdataInput(av)
	if err != nil {
		return err
	}
	symbolID, err := resolveMarketdataSymbolID(ctx, n.uow, input.SymbolID, input.SymbolCode)
	if err != nil {
		return err
	}
	input.SymbolID = symbolID
	artifact.Set(aw, resolvedMarketdataInputKey(n.id), input)
	return nil
}

type planWindowsNode struct {
	id            string
	resolveNodeID string
	chunkSizeBars int
}

func (n *planWindowsNode) Name() string { return "dagruntime.marketdata_plan_windows." + n.id }
func (n *planWindowsNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{resolvedMarketdataInputKey(n.resolveNodeID)}
}
func (n *planWindowsNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{backfillPlanKey(n.id)}
}
func (n *planWindowsNode) Reads() []state.AnyKey    { return nil }
func (n *planWindowsNode) Writes() []state.AnyKey   { return nil }
func (n *planWindowsNode) Spec() node.ExecutionSpec { return node.ExecutionSpec{Deterministic: true} }

func (n *planWindowsNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	resolved := artifact.MustGet(av, resolvedMarketdataInputKey(n.resolveNodeID))
	artifact.Set(aw, backfillPlanKey(n.id), timeframeBarBackfillPlan{
		Resolved: resolved,
		Chunks:   buildBackfillPlanChunks(resolved.TimeframeCode, resolved.From, resolved.To, n.chunkSizeBars),
	})
	return nil
}

type loadTicksForChunkNode struct {
	id         string
	planNodeID string
	uow        marketdatarepository.UnitOfWork
}

func (n *loadTicksForChunkNode) Name() string {
	return "dagruntime.marketdata_load_ticks_for_chunk." + n.id
}
func (n *loadTicksForChunkNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{backfillPlanKey(n.planNodeID)}
}
func (n *loadTicksForChunkNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{loadedBackfillPlanKey(n.id)}
}
func (n *loadTicksForChunkNode) Reads() []state.AnyKey  { return nil }
func (n *loadTicksForChunkNode) Writes() []state.AnyKey { return nil }
func (n *loadTicksForChunkNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: false}
}

func (n *loadTicksForChunkNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = txn
	plan := artifact.MustGet(av, backfillPlanKey(n.planNodeID))

	loaded := timeframeBarLoadedPlan{
		Resolved: plan.Resolved,
		Chunks:   make([]timeframeBarLoadedChunk, 0, len(plan.Chunks)),
	}
	err := n.uow.DoReadOnly(ctx, func(repos marketdatarepository.Repositories) error {
		for _, chunk := range plan.Chunks {
			ticks, err := repos.Ticks().ListBySymbolAndRange(ctx, plan.Resolved.SymbolID, chunk.From, chunk.To)
			if err != nil {
				return err
			}
			loaded.Chunks = append(loaded.Chunks, timeframeBarLoadedChunk{
				backfillChunk: chunk,
				Ticks:         ticks,
			})
		}
		return nil
	})
	if err != nil {
		return err
	}
	artifact.Set(aw, loadedBackfillPlanKey(n.id), loaded)
	return nil
}

type aggregateTimeframeBarsNode struct {
	id         string
	loadNodeID string
}

func (n *aggregateTimeframeBarsNode) Name() string {
	return "dagruntime.marketdata_aggregate_timeframe_bars." + n.id
}
func (n *aggregateTimeframeBarsNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{loadedBackfillPlanKey(n.loadNodeID)}
}
func (n *aggregateTimeframeBarsNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{aggregatedBackfillPlanKey(n.id)}
}
func (n *aggregateTimeframeBarsNode) Reads() []state.AnyKey  { return nil }
func (n *aggregateTimeframeBarsNode) Writes() []state.AnyKey { return nil }
func (n *aggregateTimeframeBarsNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}

func (n *aggregateTimeframeBarsNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	loaded := artifact.MustGet(av, loadedBackfillPlanKey(n.loadNodeID))
	aggregated := timeframeBarAggregatedPlan{
		Resolved: loaded.Resolved,
		Chunks:   make([]timeframeBarAggregatedChunk, 0, len(loaded.Chunks)),
	}
	for _, chunk := range loaded.Chunks {
		aggregated.Chunks = append(aggregated.Chunks, timeframeBarAggregatedChunk{
			backfillChunk: chunk.backfillChunk,
			Bars: timeframe.AggregateTimeframeBars(
				loaded.Resolved.TimeframeCode,
				loaded.Resolved.SymbolID,
				chunk.Ticks,
			),
		})
	}
	artifact.Set(aw, aggregatedBackfillPlanKey(n.id), aggregated)
	return nil
}

type persistTimeframeBarsNode struct {
	id              string
	planNodeID      string
	aggregateNodeID string
	mode            marketdatausecase.BackfillMode
	uow             marketdatarepository.UnitOfWork
}

func (n *persistTimeframeBarsNode) Name() string {
	return "dagruntime.marketdata_persist_timeframe_bars." + n.id
}
func (n *persistTimeframeBarsNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{
		backfillPlanKey(n.planNodeID),
		aggregatedBackfillPlanKey(n.aggregateNodeID),
	}
}
func (n *persistTimeframeBarsNode) Provides() []artifact.AnyKey {
	return []artifact.AnyKey{timeframeBarBackfillResultKey(n.id)}
}
func (n *persistTimeframeBarsNode) Reads() []state.AnyKey  { return nil }
func (n *persistTimeframeBarsNode) Writes() []state.AnyKey { return nil }
func (n *persistTimeframeBarsNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: false}
}

func (n *persistTimeframeBarsNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = txn
	plan := artifact.MustGet(av, backfillPlanKey(n.planNodeID))
	aggregated := artifact.MustGet(av, aggregatedBackfillPlanKey(n.aggregateNodeID))

	totalBars := 0
	for _, chunk := range aggregated.Chunks {
		err := n.uow.Do(ctx, func(repos marketdatarepository.Repositories) error {
			if n.mode == marketdatausecase.BackfillModeReplaceRange {
				if err := repos.TimeframeBars().DeleteBySymbolTimeframeAndRange(
					ctx,
					plan.Resolved.SymbolID,
					plan.Resolved.TimeframeCode,
					chunk.From,
					chunk.To,
				); err != nil {
					return err
				}
			}
			return repos.TimeframeBars().BulkUpsert(ctx, chunk.Bars)
		})
		if err != nil {
			return err
		}
		totalBars += len(chunk.Bars)
	}

	artifact.Set(aw, timeframeBarBackfillResultKey(n.id), marketdatausecase.BackfillTimeframeBarsResult{
		SymbolID:      plan.Resolved.SymbolID,
		TimeframeCode: plan.Resolved.TimeframeCode,
		From:          plan.Resolved.From,
		To:            plan.Resolved.To,
		ChunkCount:    len(plan.Chunks),
		BarCount:      totalBars,
	})
	return nil
}

type resolvedMarketdataInput struct {
	SymbolID      marketdata.SymbolID
	SymbolCode    string
	TimeframeCode timeframe.TimeframeCode
	From          marketdata.UTCTime
	To            marketdata.UTCTime
}

type timeframeBarBackfillPlan struct {
	Resolved resolvedMarketdataInput
	Chunks   []backfillChunk
}

type timeframeBarLoadedChunk struct {
	backfillChunk
	Ticks []marketdata.Tick
}

type timeframeBarLoadedPlan struct {
	Resolved resolvedMarketdataInput
	Chunks   []timeframeBarLoadedChunk
}

type timeframeBarAggregatedChunk struct {
	backfillChunk
	Bars []timeframe.TimeframeBar
}

type timeframeBarAggregatedPlan struct {
	Resolved resolvedMarketdataInput
	Chunks   []timeframeBarAggregatedChunk
}

type backfillChunk struct {
	From marketdata.UTCTime
	To   marketdata.UTCTime
}
