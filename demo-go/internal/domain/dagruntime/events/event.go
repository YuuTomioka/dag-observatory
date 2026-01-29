package events

import (
	"time"

	"dag-observatory/demo-go/internal/domain/dagruntime/state"
)

type Event struct {
	EventID   string
	EventTime time.Time
	Partition state.Partition
	Type      string
	Payload   any
}

