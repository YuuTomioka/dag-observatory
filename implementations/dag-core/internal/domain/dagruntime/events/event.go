package events

import (
	"time"

	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type Event struct {
	EventID   string
	EventTime time.Time
	Partition state.Partition
	Type      string
	Payload   any
}
