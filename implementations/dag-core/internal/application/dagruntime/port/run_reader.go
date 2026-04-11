package port

import (
	"time"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type RunRecord struct {
	RunID      string
	Partition  state.Partition
	EventType  string
	EventTime  time.Time
	Status     string
	StartedAt  time.Time
	EndedAt    time.Time
	Duration   time.Duration
	RetryCount int64
	Error      string
}

type RunReader interface {
	ListRuns(partition state.Partition) []RunRecord
	GetRun(runID string) (RunRecord, bool)
	ListRunSteps(runID string) []events.NodeExecutionEvent
}
