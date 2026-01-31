package usecase

import (
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

var (
	PayloadKeySymbol = events.PayloadKey[string]{
		Name:     "symbol",
		StableID: "event:input.symbol.v1",
	}
	PayloadKeyMode = events.PayloadKey[string]{
		Name:     "mode",
		StableID: "event:input.mode.v1",
	}

	InputKeySymbol = artifact.Key[string]{Name: "symbol", StableID: "artifact:input.symbol.v1"}
	InputKeyMode   = artifact.Key[string]{Name: "mode", StableID: "artifact:input.mode.v1"}
)
