package di

import (
	"context"
	"testing"

	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
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
