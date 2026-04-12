package usecase

import (
	"context"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type fakeRunReader struct {
	runs  []port.RunRecord
	steps map[string][]events.NodeExecutionEvent
}

func (r *fakeRunReader) ListRuns(partition state.Partition) []port.RunRecord {
	out := make([]port.RunRecord, 0, len(r.runs))
	for _, run := range r.runs {
		if partition != "" && run.Partition != partition {
			continue
		}
		out = append(out, run)
	}
	return out
}

func (r *fakeRunReader) GetRun(runID string) (port.RunRecord, bool) {
	for _, run := range r.runs {
		if run.RunID == runID {
			return run, true
		}
	}
	return port.RunRecord{}, false
}

func (r *fakeRunReader) ListRunSteps(runID string) []events.NodeExecutionEvent {
	items := r.steps[runID]
	out := make([]events.NodeExecutionEvent, len(items))
	copy(out, items)
	return out
}

func TestListRunsExecute(t *testing.T) {
	now := time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)
	reader := &fakeRunReader{
		runs: []port.RunRecord{
			{
				RunID:      "run-1",
				Partition:  state.Partition("p1"),
				Status:     "succeeded",
				EventType:  "task.completed",
				EventTime:  now,
				StartedAt:  now,
				EndedAt:    now.Add(time.Second),
				Duration:   time.Second,
				RetryCount: 1,
			},
		},
		steps: map[string][]events.NodeExecutionEvent{
			"run-1": {{SequenceNo: 1}, {SequenceNo: 2}},
		},
	}
	uc := &ListRuns{Reader: reader}

	items, err := uc.Execute(context.Background(), ListRunsRequest{Partition: "p1"})
	if err != nil {
		t.Fatalf("list runs failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one run, got %d", len(items))
	}
	if items[0].RunID != "run-1" || items[0].StepCount != 2 {
		t.Fatalf("unexpected run view: %#v", items[0])
	}
}

func TestGetRunNodeExecute(t *testing.T) {
	reader := &fakeRunReader{
		steps: map[string][]events.NodeExecutionEvent{
			"run-1": {
				{SequenceNo: 1, NodeName: "n1", Status: events.NodeExecutionStatusSucceeded},
				{SequenceNo: 2, NodeName: "n2", Status: events.NodeExecutionStatusFailed},
			},
		},
	}
	uc := &GetRunNode{Reader: reader}

	item, ok, err := uc.Execute(context.Background(), GetRunNodeRequest{
		RunID:      "run-1",
		SequenceNo: "2",
	})
	if err != nil {
		t.Fatalf("get run node failed: %v", err)
	}
	if !ok {
		t.Fatal("expected run node to exist")
	}
	if item.NodeName != "n2" || item.NodeExecutionID != "2" {
		t.Fatalf("unexpected run step view: %#v", item)
	}
}

func TestCompareRunsExecute(t *testing.T) {
	now := time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)
	reader := &fakeRunReader{
		runs: []port.RunRecord{
			{
				RunID:      "run-base",
				Partition:  state.Partition("p1"),
				Status:     "succeeded",
				EventType:  "task.completed",
				EventTime:  now,
				StartedAt:  now,
				EndedAt:    now.Add(2 * time.Second),
				Duration:   2 * time.Second,
				RetryCount: 0,
			},
			{
				RunID:      "run-target",
				Partition:  state.Partition("p1"),
				Status:     "failed",
				EventType:  "task.failed",
				EventTime:  now.Add(10 * time.Second),
				StartedAt:  now.Add(10 * time.Second),
				EndedAt:    now.Add(14 * time.Second),
				Duration:   4 * time.Second,
				RetryCount: 1,
			},
		},
		steps: map[string][]events.NodeExecutionEvent{
			"run-base": {
				{
					SequenceNo: 1,
					Status:     events.NodeExecutionStatusSucceeded,
					StateDiff: []events.StateDiffField{
						{Field: "strategy_summary.trade_count", After: 1.0},
					},
				},
			},
			"run-target": {
				{
					SequenceNo: 1,
					Status:     events.NodeExecutionStatusFailed,
					StateDiff: []events.StateDiffField{
						{Field: "strategy_summary.trade_count", After: 3.0},
					},
				},
			},
		},
	}

	uc := &CompareRuns{Reader: reader}
	view, ok, err := uc.Execute(context.Background(), CompareRunsRequest{
		BaseRunID:   "run-base",
		TargetRunID: "run-target",
	})
	if err != nil {
		t.Fatalf("compare runs failed: %v", err)
	}
	if !ok {
		t.Fatal("expected compare view to exist")
	}
	if !view.SummaryDelta.StatusChanged {
		t.Fatalf("expected status_changed=true, got %#v", view.SummaryDelta)
	}
	if view.SummaryDelta.DurationMSDelta != 2000 {
		t.Fatalf("expected duration delta 2000ms, got %d", view.SummaryDelta.DurationMSDelta)
	}
	if len(view.StateFieldDeltas) != 1 {
		t.Fatalf("expected 1 state field delta, got %d", len(view.StateFieldDeltas))
	}
	if view.StateFieldDeltas[0].Field != "strategy_summary.trade_count" || !view.StateFieldDeltas[0].Changed {
		t.Fatalf("unexpected state field delta: %#v", view.StateFieldDeltas[0])
	}
}
