package recorder

import (
	"context"
	"sort"
	"strings"
	"sync"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type InMemoryRecorder struct {
	mu    sync.RWMutex
	runs  map[string]port.RunRecord
	steps map[string][]events.NodeExecutionEvent
}

func NewInMemoryRecorder() *InMemoryRecorder {
	return &InMemoryRecorder{
		runs:  map[string]port.RunRecord{},
		steps: map[string][]events.NodeExecutionEvent{},
	}
}

func (r *InMemoryRecorder) RecordEvent(ctx context.Context, event events.Event) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()

	record, ok := r.runs[event.EventID]
	if !ok {
		record = port.RunRecord{
			RunID: event.EventID,
		}
	}
	record.Partition = event.Partition
	record.EventType = event.Type
	record.EventTime = event.EventTime
	record.StartedAt = event.EventTime
	record.Status = normalizeRunStatus(event.Type)
	r.runs[event.EventID] = record
}

func (r *InMemoryRecorder) RecordNodeExecution(ctx context.Context, event events.NodeExecutionEvent) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	runID := event.RunID
	if runID == "" {
		return
	}
	r.steps[runID] = append(r.steps[runID], event)
}

func (r *InMemoryRecorder) RecordNodeResult(ctx context.Context, result port.NodeResult) {
	_ = ctx
	_ = result
}

func (r *InMemoryRecorder) RecordCycleResult(ctx context.Context, result port.CycleResult) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	runID := result.Event.EventID
	if runID == "" {
		return
	}
	record, ok := r.runs[runID]
	if !ok {
		record = port.RunRecord{
			RunID: runID,
		}
	}
	record.Partition = result.Partition
	record.EventType = result.Event.Type
	record.EventTime = result.Event.EventTime
	if record.StartedAt.IsZero() {
		record.StartedAt = result.Event.EventTime
	}
	record.EndedAt = result.Event.EventTime.Add(result.Duration)
	record.Duration = result.Duration
	record.RetryCount = result.RetryCount
	if result.Err != nil {
		record.Status = "failed"
		record.Error = result.Err.Error()
	} else {
		record.Status = "succeeded"
		record.Error = ""
	}
	r.runs[runID] = record
}

func (r *InMemoryRecorder) ListRuns(partition state.Partition) []port.RunRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]port.RunRecord, 0, len(r.runs))
	for _, run := range r.runs {
		if partition != "" && run.Partition != partition {
			continue
		}
		out = append(out, run)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].StartedAt.Equal(out[j].StartedAt) {
			return out[i].RunID > out[j].RunID
		}
		return out[i].StartedAt.After(out[j].StartedAt)
	})
	return out
}

func (r *InMemoryRecorder) GetRun(runID string) (port.RunRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	run, ok := r.runs[runID]
	return run, ok
}

func (r *InMemoryRecorder) ListRunSteps(runID string) []events.NodeExecutionEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := r.steps[runID]
	out := make([]events.NodeExecutionEvent, len(items))
	copy(out, items)
	sort.Slice(out, func(i, j int) bool {
		return out[i].SequenceNo < out[j].SequenceNo
	})
	return out
}

func normalizeRunStatus(eventType string) string {
	parts := strings.Split(eventType, ".")
	if len(parts) == 0 {
		return "unknown"
	}
	switch parts[len(parts)-1] {
	case "requested", "started", "running":
		return "in_progress"
	case "completed", "succeeded":
		return "succeeded"
	case "failed", "error":
		return "failed"
	case "cancelled", "canceled":
		return "cancelled"
	default:
		return "unknown"
	}
}
