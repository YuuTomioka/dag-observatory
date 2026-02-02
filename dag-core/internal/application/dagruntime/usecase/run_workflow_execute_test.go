package usecase

import (
	"context"
	"testing"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/driver"
	"dag-observatory/dag-core/internal/domain/dagruntime/engine"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"
	"dag-observatory/dag-core/internal/domain/dagruntime/policy"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	artifactinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
)

type recordingEnqueuer struct {
	last events.Event
}

func (r *recordingEnqueuer) Enqueue(ctx context.Context, event events.Event) error {
	_ = ctx
	r.last = event
	return nil
}

type recordingRecorder struct {
	last events.Event
}

func (r *recordingRecorder) RecordEvent(ctx context.Context, event events.Event) {
	_ = ctx
	r.last = event
}

func (r *recordingRecorder) RecordNodeResult(ctx context.Context, result port.NodeResult) {
	_ = ctx
	_ = result
}

func (r *recordingRecorder) RecordCycleResult(ctx context.Context, result port.CycleResult) {
	_ = ctx
	_ = result
}

func TestBuildRunEventDefaults(t *testing.T) {
	result, event, partition := buildRunEvent(RunWorkflowRequest{})

	if result.RunID == "" {
		t.Fatal("expected run_id to be generated")
	}
	if result.Symbol != defaultSymbol {
		t.Fatalf("expected default symbol %q, got %q", defaultSymbol, result.Symbol)
	}
	if result.Mode != defaultMode {
		t.Fatalf("expected default mode %q, got %q", defaultMode, result.Mode)
	}
	if event.EventID != result.RunID {
		t.Fatalf("expected event id %q, got %q", result.RunID, event.EventID)
	}
	if partition != state.Partition(result.RunID) {
		t.Fatalf("expected partition %q, got %q", result.RunID, partition)
	}

	payload, ok := event.Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected payload map, got %T", event.Payload)
	}
	input, ok := payload["input"].(map[string]any)
	if !ok {
		t.Fatalf("expected input map, got %T", payload["input"])
	}
	if input["symbol"] != defaultSymbol {
		t.Fatalf("expected input.symbol %q, got %v", defaultSymbol, input["symbol"])
	}
	if input["mode"] != defaultMode {
		t.Fatalf("expected input.mode %q, got %v", defaultMode, input["mode"])
	}
	if payload["symbol"] != defaultSymbol {
		t.Fatalf("expected payload.symbol %q, got %v", defaultSymbol, payload["symbol"])
	}
	if payload["mode"] != defaultMode {
		t.Fatalf("expected payload.mode %q, got %v", defaultMode, payload["mode"])
	}
}

func TestBuildRunEventWithRunID(t *testing.T) {
	req := RunWorkflowRequest{RunID: "fixed-id", Symbol: "USDJPY", Mode: "normal"}
	result, event, partition := buildRunEvent(req)

	if result.RunID != "fixed-id" {
		t.Fatalf("expected run_id fixed-id, got %q", result.RunID)
	}
	if event.EventID != "fixed-id" {
		t.Fatalf("expected event id fixed-id, got %q", event.EventID)
	}
	if partition != state.Partition("fixed-id") {
		t.Fatalf("expected partition fixed-id, got %q", partition)
	}
}

func TestExecuteEnqueuePayloadIsEnvelope(t *testing.T) {
	enqueuer := &recordingEnqueuer{}
	uc := &RunWorkflow{Enqueuer: enqueuer}

	result, err := uc.Execute(context.Background(), RunWorkflowRequest{
		Symbol: "EURUSD",
		Mode:   "debug",
	})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if !result.EnqueueMode {
		t.Fatal("expected enqueue mode result")
	}
	if _, ok := enqueuer.last.Payload.([]events.PayloadEnvelope); !ok {
		t.Fatalf("expected envelope payload, got %T", enqueuer.last.Payload)
	}
}

func TestExecuteDriverPayloadIsInputMap(t *testing.T) {
	recorder := &recordingRecorder{}
	runner := &engine.Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{DefaultRetry: policy.RetryPolicy{MaxAttempts: 1}},
		Recorder:      recorder,
	}
	drv := &driver.Driver{
		Runner:   runner,
		Compiled: pipeline.Compiled{Name: "test"},
	}
	uc := &RunWorkflow{Driver: drv}

	result, err := uc.Execute(context.Background(), RunWorkflowRequest{
		Symbol: "EURUSD",
		Mode:   "debug",
	})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if result.EnqueueMode {
		t.Fatal("expected driver mode result")
	}
	if _, ok := recorder.last.Payload.(engine.InputMap); !ok {
		t.Fatalf("expected input map payload, got %T", recorder.last.Payload)
	}
}
