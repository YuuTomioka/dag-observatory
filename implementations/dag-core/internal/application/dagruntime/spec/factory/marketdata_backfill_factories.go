package factory

import (
	"fmt"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	marketdatarepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
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
