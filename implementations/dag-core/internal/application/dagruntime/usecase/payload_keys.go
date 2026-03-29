package usecase

import (
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

var (
	PayloadKeyRunID = events.PayloadKey[string]{
		Name:     "run_id",
		StableID: "event:run_id.v1",
	}
	PayloadKeyTaskID = events.PayloadKey[string]{
		Name:     "task_id",
		StableID: "event:task_id.v1",
	}
	PayloadKeyAttempt = events.PayloadKey[int]{
		Name:     "attempt",
		StableID: "event:attempt.v1",
	}
	PayloadKeyTaskName = events.PayloadKey[string]{
		Name:     "task_name",
		StableID: "event:task_name.v1",
	}
	PayloadKeyInput = events.PayloadKey[map[string]any]{
		Name:     "input",
		StableID: "event:input.v1",
	}
	PayloadKeyOutput = events.PayloadKey[map[string]any]{
		Name:     "output",
		StableID: "event:output.v1",
	}
	PayloadKeyError = events.PayloadKey[map[string]any]{
		Name:     "error",
		StableID: "event:error.v1",
	}

	PayloadKeySymbol = events.PayloadKey[string]{
		Name:     "symbol",
		StableID: "event:input.symbol.v1",
	}
	PayloadKeyMode = events.PayloadKey[string]{
		Name:     "mode",
		StableID: "event:input.mode.v1",
	}
	PayloadKeyMarketBars = events.PayloadKey[[]float64]{
		Name:     "market_bars",
		StableID: "event:input.market.bars.v1",
	}

	InputKeySymbol     = artifact.Key[string]{Name: "symbol", StableID: "artifact:input.symbol.v1"}
	InputKeyMode       = artifact.Key[string]{Name: "mode", StableID: "artifact:input.mode.v1"}
	InputKeyMarketBars = artifact.Key[[]float64]{Name: "market.bars", StableID: "artifact:input.market.bars.v1"}
)
