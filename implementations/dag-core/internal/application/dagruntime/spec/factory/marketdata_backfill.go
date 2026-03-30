package factory

import (
	"context"
	"fmt"

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

func timeframeBarBackfillResultKey(nodeID string) artifact.Key[marketdatausecase.BackfillTimeframeBarsResult] {
	return artifact.Key[marketdatausecase.BackfillTimeframeBarsResult]{
		Name:     fmt.Sprintf("%s.backfill_result", nodeID),
		StableID: fmt.Sprintf("artifact:dagruntime.marketdata.backfill.%s.v1", nodeID),
	}
}
