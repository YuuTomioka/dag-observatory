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

func TestReplayJSONLReconstructsRunReadModel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runs.jsonl")
	primary := NewInMemoryRecorder()
	writer, err := NewJSONLRecorder(path, primary)
	if err != nil {
		t.Fatalf("new jsonl recorder: %v", err)
	}

	event := events.Event{
		EventID:   "run-replay-1",
		EventTime: time.Date(2026, 4, 12, 5, 0, 0, 0, time.UTC),
		Partition: state.Partition("partition-replay"),
		Type:      "task.requested",
	}
	writer.RecordEvent(context.Background(), event)
	writer.RecordNodeExecution(context.Background(), events.NodeExecutionEvent{
		RunID:      "run-replay-1",
		Partition:  state.Partition("partition-replay"),
		EventTime:  event.EventTime,
		SequenceNo: 1,
		NodeID:     "node-1",
		NodeName:   "node.one",
		Status:     events.NodeExecutionStatusSucceeded,
	})
	writer.RecordCycleResult(context.Background(), port.CycleResult{
		Partition: state.Partition("partition-replay"),
		Event:     event,
		Duration:  5 * time.Millisecond,
	})
	if err := writer.Close(); err != nil {
		t.Fatalf("close jsonl recorder: %v", err)
	}

	replayed := NewInMemoryRecorder()
	stats, err := ReplayJSONL(path, replayed)
	if err != nil {
		t.Fatalf("replay jsonl: %v", err)
	}
	if stats.AppliedLines != 3 || stats.SkippedLines != 0 {
		t.Fatalf("unexpected replay stats: %#v", stats)
	}

	run, ok := replayed.GetRun("run-replay-1")
	if !ok {
		t.Fatal("expected replayed run")
	}
	if run.Status != "succeeded" {
		t.Fatalf("expected succeeded status, got %q", run.Status)
	}
	steps := replayed.ListRunSteps("run-replay-1")
	if len(steps) != 1 || steps[0].NodeID != "node-1" {
		t.Fatalf("unexpected replayed steps: %#v", steps)
	}
}

func TestReplayJSONLToleratesMalformedLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runs-malformed.jsonl")
	content := strings.Join([]string{
		`{"kind":"run_event","run_id":"run-x","partition":"p","event_type":"task.requested","event_time":"2026-04-12T08:00:00Z"}`,
		`{"kind":`,
		`{"kind":"node_execution","run_id":"run-x","partition":"p","event_time":"2026-04-12T08:00:00Z","sequence_no":1,"node_id":"n1","node_name":"node.1","status":"succeeded"}`,
		`{"kind":"cycle_result","run_id":"run-x","partition":"p","event_type":"task.requested","event_time":"2026-04-12T08:00:00Z","duration_ns":1000000}`,
	}, "\n")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write malformed jsonl: %v", err)
	}

	replayed := NewInMemoryRecorder()
	stats, err := ReplayJSONL(path, replayed)
	if err != nil {
		t.Fatalf("replay jsonl should tolerate malformed lines: %v", err)
	}
	if stats.DecodeErrors != 1 || stats.SkippedLines != 1 {
		t.Fatalf("unexpected replay stats: %#v", stats)
	}
	run, ok := replayed.GetRun("run-x")
	if !ok || run.Status != "succeeded" {
		t.Fatalf("expected valid trailing records to be replayed, run=%#v ok=%v", run, ok)
	}
}
