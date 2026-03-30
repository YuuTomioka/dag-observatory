package factory

import (
	"context"
	"fmt"
	"strings"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	marketdatarepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

type TimeframeBarBackfillFactory struct {
	UnitOfWork marketdatarepository.UnitOfWork
}

func (f *TimeframeBarBackfillFactory) Kind() string { return "marketdata_timeframe_bar_backfill" }

func (f *TimeframeBarBackfillFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if f == nil || f.UnitOfWork == nil {
		return nil, fmt.Errorf("marketdata_timeframe_bar_backfill: marketdata unit_of_work is not configured")
	}
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"mode", "chunk_size_bars"}); err != nil {
		return nil, fmt.Errorf("marketdata_timeframe_bar_backfill node %q: %w", nodeSpec.ID, err)
	}
	mode := optionalString(nodeSpec.Config, "mode", string(marketdatausecase.BackfillModeReplaceRange))
	chunkSizeBars := optionalInt(nodeSpec.Config, "chunk_size_bars", 1000)
	if chunkSizeBars <= 0 {
		return nil, fmt.Errorf("marketdata_timeframe_bar_backfill node %q: config.chunk_size_bars must be > 0", nodeSpec.ID)
	}
	return &timeframeBarBackfillNode{
		id: nodeSpec.ID,
		usecase: &marketdatausecase.BackfillTimeframeBars{
			UnitOfWork: f.UnitOfWork,
		},
		mode:          marketdatausecase.BackfillMode(mode),
		chunkSizeBars: chunkSizeBars,
	}, nil
}

type ResolveSymbolFactory struct {
	UnitOfWork marketdatarepository.UnitOfWork
}

func (f *ResolveSymbolFactory) Kind() string { return "marketdata_resolve_symbol" }

func (f *ResolveSymbolFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if f == nil || f.UnitOfWork == nil {
		return nil, fmt.Errorf("marketdata_resolve_symbol: marketdata unit_of_work is not configured")
	}
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, nil); err != nil {
		return nil, fmt.Errorf("marketdata_resolve_symbol node %q: %w", nodeSpec.ID, err)
	}
	return &resolveSymbolNode{id: nodeSpec.ID, uow: f.UnitOfWork}, nil
}

type PlanWindowsFactory struct{}

func (f *PlanWindowsFactory) Kind() string { return "marketdata_plan_windows" }

func (f *PlanWindowsFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"resolve_node_id", "chunk_size_bars"}); err != nil {
		return nil, fmt.Errorf("marketdata_plan_windows node %q: %w", nodeSpec.ID, err)
	}
	resolveNodeID, err := requiredString(nodeSpec.Config, "resolve_node_id")
	if err != nil {
		return nil, fmt.Errorf("marketdata_plan_windows node %q: %w", nodeSpec.ID, err)
	}
	chunkSizeBars := optionalInt(nodeSpec.Config, "chunk_size_bars", 1000)
	if chunkSizeBars <= 0 {
		return nil, fmt.Errorf("marketdata_plan_windows node %q: config.chunk_size_bars must be > 0", nodeSpec.ID)
	}
	return &planWindowsNode{
		id:            nodeSpec.ID,
		resolveNodeID: resolveNodeID,
		chunkSizeBars: chunkSizeBars,
	}, nil
}

type LoadTicksForChunkFactory struct {
	UnitOfWork marketdatarepository.UnitOfWork
}

func (f *LoadTicksForChunkFactory) Kind() string { return "marketdata_load_ticks_for_chunk" }

func (f *LoadTicksForChunkFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if f == nil || f.UnitOfWork == nil {
		return nil, fmt.Errorf("marketdata_load_ticks_for_chunk: marketdata unit_of_work is not configured")
	}
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"plan_node_id"}); err != nil {
		return nil, fmt.Errorf("marketdata_load_ticks_for_chunk node %q: %w", nodeSpec.ID, err)
	}
	planNodeID, err := requiredString(nodeSpec.Config, "plan_node_id")
	if err != nil {
		return nil, fmt.Errorf("marketdata_load_ticks_for_chunk node %q: %w", nodeSpec.ID, err)
	}
	return &loadTicksForChunkNode{
		id:         nodeSpec.ID,
		planNodeID: planNodeID,
		uow:        f.UnitOfWork,
	}, nil
}

type AggregateTimeframeBarsFactory struct{}

func (f *AggregateTimeframeBarsFactory) Kind() string { return "marketdata_aggregate_timeframe_bars" }

func (f *AggregateTimeframeBarsFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"load_node_id"}); err != nil {
		return nil, fmt.Errorf("marketdata_aggregate_timeframe_bars node %q: %w", nodeSpec.ID, err)
	}
	loadNodeID, err := requiredString(nodeSpec.Config, "load_node_id")
	if err != nil {
		return nil, fmt.Errorf("marketdata_aggregate_timeframe_bars node %q: %w", nodeSpec.ID, err)
	}
	return &aggregateTimeframeBarsNode{
		id:         nodeSpec.ID,
		loadNodeID: loadNodeID,
	}, nil
}

type PersistTimeframeBarsFactory struct {
	UnitOfWork marketdatarepository.UnitOfWork
}

func (f *PersistTimeframeBarsFactory) Kind() string { return "marketdata_persist_timeframe_bars" }

func (f *PersistTimeframeBarsFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if f == nil || f.UnitOfWork == nil {
		return nil, fmt.Errorf("marketdata_persist_timeframe_bars: marketdata unit_of_work is not configured")
	}
	if err := ensureNoUnknownConfigKeys(nodeSpec.Config, []string{"plan_node_id", "aggregate_node_id", "mode"}); err != nil {
		return nil, fmt.Errorf("marketdata_persist_timeframe_bars node %q: %w", nodeSpec.ID, err)
	}
	planNodeID, err := requiredString(nodeSpec.Config, "plan_node_id")
	if err != nil {
		return nil, fmt.Errorf("marketdata_persist_timeframe_bars node %q: %w", nodeSpec.ID, err)
	}
	aggregateNodeID, err := requiredString(nodeSpec.Config, "aggregate_node_id")
	if err != nil {
		return nil, fmt.Errorf("marketdata_persist_timeframe_bars node %q: %w", nodeSpec.ID, err)
	}
	mode := optionalString(nodeSpec.Config, "mode", string(marketdatausecase.BackfillModeReplaceRange))
	switch marketdatausecase.BackfillMode(mode) {
	case marketdatausecase.BackfillModeReplaceRange, marketdatausecase.BackfillModeUpsertOnly:
	default:
		return nil, fmt.Errorf("marketdata_persist_timeframe_bars node %q: config.mode must be replace_range or upsert_only", nodeSpec.ID)
	}
	return &persistTimeframeBarsNode{
		id:              nodeSpec.ID,
		planNodeID:      planNodeID,
		aggregateNodeID: aggregateNodeID,
		mode:            marketdatausecase.BackfillMode(mode),
		uow:             f.UnitOfWork,
	}, nil
}

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
		TimeframeCode: marketdata.TimeframeCode(artifact.MustGet(av, dagruntimeusecase.InputKeyMarketdataTimeframeCode)),
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
			Bars: marketdata.AggregateTimeframeBars(
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
	TimeframeCode marketdata.TimeframeCode
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
	Bars []marketdata.TimeframeBar
}

type timeframeBarAggregatedPlan struct {
	Resolved resolvedMarketdataInput
	Chunks   []timeframeBarAggregatedChunk
}

type backfillChunk struct {
	From marketdata.UTCTime
	To   marketdata.UTCTime
}

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
