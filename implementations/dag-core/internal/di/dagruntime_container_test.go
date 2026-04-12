package di

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	recorderinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/recorder"
)

func TestNewDAGRuntimeContainerMemoryModeRunsInProcess(t *testing.T) {
	t.Parallel()

	cfg := Config{
		ServiceName:      "dag-observatory-dag-core",
		StateStoreType:   "memory",
		EventStoreType:   "memory",
		WorkflowSpecPath: "workflows/default.yaml",
	}
	compiled, err := compileDefaultWorkflow(cfg, nil)
	if err != nil {
		t.Fatalf("compile default workflow: %v", err)
	}

	container, err := NewDAGRuntimeContainer(cfg, nil, compiled)
	if err != nil {
		t.Fatalf("new dagruntime container: %v", err)
	}

	result, err := container.Usecase.Execute(context.Background(), usecase.RunWorkflowRequest{
		Symbol: "USDJPY",
		Mode:   "normal",
	})
	if err != nil {
		t.Fatalf("execute workflow: %v", err)
	}
	if result.EnqueueMode {
		t.Fatal("expected in-process driver mode in memory event store")
	}
	if container.RunReader == nil {
		t.Fatal("expected run reader to be configured")
	}
	record, ok := container.RunReader.GetRun(result.RunID)
	if !ok {
		t.Fatalf("expected recorded run for run_id=%s", result.RunID)
	}
	if record.Status != "succeeded" {
		t.Fatalf("expected succeeded run status, got %q", record.Status)
	}
}

func TestNewDAGRuntimeContainerReplaysObservationJSONLAtStartup(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "runs.jsonl")
	memory := recorderinfra.NewInMemoryRecorder()
	writer, err := recorderinfra.NewJSONLRecorder(path, memory)
	if err != nil {
		t.Fatalf("new jsonl recorder: %v", err)
	}
	event := events.Event{
		EventID:   "run-replayed",
		EventTime: time.Date(2026, 4, 12, 7, 0, 0, 0, time.UTC),
		Partition: state.Partition("partition-replayed"),
		Type:      "task.requested",
	}
	writer.RecordEvent(context.Background(), event)
	writer.RecordNodeExecution(context.Background(), events.NodeExecutionEvent{
		RunID:      "run-replayed",
		Partition:  state.Partition("partition-replayed"),
		EventTime:  event.EventTime,
		SequenceNo: 1,
		NodeID:     "node-A",
		NodeName:   "node.a",
		Status:     events.NodeExecutionStatusSucceeded,
	})
	writer.RecordCycleResult(context.Background(), port.CycleResult{
		Partition: state.Partition("partition-replayed"),
		Event:     event,
		Duration:  3 * time.Millisecond,
	})
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	cfg := Config{
		ServiceName:          "dag-observatory-dag-core",
		StateStoreType:       "memory",
		EventStoreType:       "memory",
		WorkflowSpecPath:     "workflows/default.yaml",
		ObservationJSONLPath: path,
	}
	compiled, err := compileDefaultWorkflow(cfg, nil)
	if err != nil {
		t.Fatalf("compile default workflow: %v", err)
	}

	container, err := NewDAGRuntimeContainer(cfg, nil, compiled)
	if err != nil {
		t.Fatalf("new dagruntime container: %v", err)
	}
	if container.ReplayStats.AppliedLines != 3 || container.ReplayStats.SkippedLines != 0 {
		t.Fatalf("unexpected replay stats: %#v", container.ReplayStats)
	}
	run, ok := container.RunReader.GetRun("run-replayed")
	if !ok {
		t.Fatal("expected replayed run in read model")
	}
	if run.Partition != state.Partition("partition-replayed") {
		t.Fatalf("unexpected partition: %s", run.Partition)
	}
	steps := container.RunReader.ListRunSteps("run-replayed")
	if len(steps) != 1 || steps[0].SequenceNo != 1 {
		t.Fatalf("unexpected replayed steps: %#v", steps)
	}
}
