package recorder

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

func TestJSONLRecorderAppendsObservationRecords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runs.jsonl")
	next := NewInMemoryRecorder()
	recorder, err := NewJSONLRecorder(path, next)
	if err != nil {
		t.Fatalf("new jsonl recorder: %v", err)
	}
	defer recorder.Close()

	event := events.Event{
		EventID:   "run-1",
		EventTime: time.Date(2026, 4, 12, 3, 0, 0, 0, time.UTC),
		Partition: state.Partition("partition-a"),
		Type:      "task.requested",
	}
	recorder.RecordEvent(context.Background(), event)
	recorder.RecordNodeExecution(context.Background(), events.NodeExecutionEvent{
		RunID:       "run-1",
		Partition:   state.Partition("partition-a"),
		EventTime:   event.EventTime,
		SequenceNo:  1,
		IntentID:    "intent:1",
		ExecutionID: "exec:1",
		TradeID:     "trade:1",
		NodeID:      "n1",
		NodeName:    "node.one",
		Status:      events.NodeExecutionStatusSucceeded,
		DurationNS:  1234,
	})
	recorder.RecordCycleResult(context.Background(), port.CycleResult{
		Partition: state.Partition("partition-a"),
		Event:     event,
		Duration:  2 * time.Millisecond,
	})

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read jsonl file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	var records []jsonlObservationRecord
	for _, line := range lines {
		var record jsonlObservationRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("decode line %q: %v", line, err)
		}
		records = append(records, record)
	}
	if records[0].Kind != "run_event" || records[1].Kind != "node_execution" || records[2].Kind != "cycle_result" {
		t.Fatalf("unexpected record kinds: %#v", records)
	}
	if records[1].IntentID != "intent:1" || records[1].ExecutionID != "exec:1" || records[1].TradeID != "trade:1" {
		t.Fatalf("unexpected node correlation ids: %#v", records[1])
	}
	if len(next.ListRunSteps("run-1")) != 1 {
		t.Fatalf("expected wrapped recorder to keep in-memory steps")
	}
}
