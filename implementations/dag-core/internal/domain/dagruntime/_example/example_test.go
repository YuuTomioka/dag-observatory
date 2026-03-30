package _example

import (
	"context"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/engine"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"
	"dag-observatory/dag-core/internal/domain/dagruntime/policy"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/dagruntime/workflow"
	artifactinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
)

func TestDeterministicRun(t *testing.T) {
	compiled := mustCompile(t, BuildWorkflow(false))

	output1 := runOnce(t, compiled, "USDJPY", "hello")
	output2 := runOnce(t, compiled, "USDJPY", "hello")

	if output1 != output2 {
		t.Fatalf("expected deterministic output, got %q and %q", output1, output2)
	}
}

func TestRollbackOnFailure(t *testing.T) {
	compiled := mustCompile(t, BuildWorkflow(true))

	artifactStore := artifactinfra.NewMemoryStore()
	stateStore := stateinfra.NewMemoryStore()
	runner := &engine.Runner{
		ArtifactStore: artifactStore,
		StateStore:    stateStore,
		Policy:        policy.Policy{},
	}

	inputs := engine.InputMap{
		KeyInput: "hello",
	}
	event := events.Event{
		EventID:   "run-fail",
		EventTime: time.Now(),
		Partition: state.Partition("USDJPY"),
		Type:      "test",
	}

	if err := runner.RunCycle(context.Background(), compiled, inputs, event.Partition, event); err == nil {
		t.Fatalf("expected error, got nil")
	}

	txn := stateStore.BeginTxn(event.Partition)
	if _, ok := state.Get(txn, StateCount); ok {
		t.Fatalf("expected rollback, state should not be written")
	}
	txn.Rollback()
}

func TestPartitionIsolation(t *testing.T) {
	compiled := mustCompile(t, BuildWorkflow(false))

	artifactStore := artifactinfra.NewMemoryStore()
	stateStore := stateinfra.NewMemoryStore()
	runner := &engine.Runner{
		ArtifactStore: artifactStore,
		StateStore:    stateStore,
		Policy:        policy.Policy{},
	}

	runWithPartition(t, runner, compiled, "USDJPY", "hello")
	runWithPartition(t, runner, compiled, "EURUSD", "hello")

	checkCount(t, stateStore, state.Partition("USDJPY"), 1)
	checkCount(t, stateStore, state.Partition("EURUSD"), 1)
}

func runOnce(t *testing.T, compiled pipeline.Compiled, symbol, input string) string {
	t.Helper()

	artifactStore := artifactinfra.NewMemoryStore()
	stateStore := stateinfra.NewMemoryStore()
	runner := &engine.Runner{
		ArtifactStore: artifactStore,
		StateStore:    stateStore,
		Policy:        policy.Policy{},
	}

	inputs := engine.InputMap{
		KeyInput: input,
	}
	event := events.Event{
		EventID:   "run",
		EventTime: time.Now(),
		Partition: state.Partition(symbol),
		Type:      "test",
	}

	if err := runner.RunCycle(context.Background(), compiled, inputs, event.Partition, event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	value, ok := artifact.Get(artifactStore.View(), KeyResult)
	if !ok {
		t.Fatalf("expected result artifact")
	}
	return value
}

func runWithPartition(t *testing.T, runner *engine.Runner, compiled pipeline.Compiled, symbol, input string) {
	t.Helper()

	inputs := engine.InputMap{
		KeyInput: input,
	}
	event := events.Event{
		EventID:   "run-" + symbol,
		EventTime: time.Now(),
		Partition: state.Partition(symbol),
		Type:      "test",
	}

	if err := runner.RunCycle(context.Background(), compiled, inputs, event.Partition, event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func checkCount(t *testing.T, store *stateinfra.MemoryStore, partition state.Partition, expected int) {
	t.Helper()
	txn := store.BeginTxn(partition)
	value, ok := state.Get(txn, StateCount)
	if !ok {
		t.Fatalf("expected state count for %s", partition)
	}
	if value != expected {
		t.Fatalf("expected count %d, got %d", expected, value)
	}
	txn.Rollback()
}

func mustCompile(t *testing.T, wf workflow.Workflow) pipeline.Compiled {
	t.Helper()
	compiled, err := workflow.Compile(wf)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	return compiled
}
