package usecase

import (
	"context"
	"testing"
	"time"

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

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
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
	now := time.Date(2026, 3, 29, 8, 0, 0, 0, time.UTC)
	result, event, partition := buildRunEvent(RunWorkflowRequest{}, now)

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
	if !event.EventTime.Equal(now) {
		t.Fatalf("expected event time %s, got %s", now, event.EventTime)
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
	now := time.Date(2026, 3, 29, 9, 0, 0, 0, time.UTC)
	result, event, partition := buildRunEvent(req, now)

	if result.RunID != "fixed-id" {
		t.Fatalf("expected run_id fixed-id, got %q", result.RunID)
	}
	if event.EventID != "fixed-id" {
		t.Fatalf("expected event id fixed-id, got %q", event.EventID)
	}
	if partition != state.Partition("fixed-id") {
		t.Fatalf("expected partition fixed-id, got %q", partition)
	}
	if !event.EventTime.Equal(now) {
		t.Fatalf("expected event time %s, got %s", now, event.EventTime)
	}
}

func TestBuildRunEventWithMarketBars(t *testing.T) {
	req := RunWorkflowRequest{
		RunID:  "bars-run",
		Symbol: "USDJPY",
		Mode:   "normal",
		Bars:   []float64{1.0, 2.0, 3.0},
	}
	now := time.Date(2026, 3, 29, 9, 30, 0, 0, time.UTC)
	_, event, _ := buildRunEvent(req, now)

	payload, ok := event.Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected payload map, got %T", event.Payload)
	}
	if _, ok := payload["market_bars"]; !ok {
		t.Fatal("expected market_bars in payload")
	}
	input, ok := payload["input"].(map[string]any)
	if !ok {
		t.Fatalf("expected input map, got %T", payload["input"])
	}
	if _, ok := input["bars"]; !ok {
		t.Fatal("expected bars in input payload")
	}
}

func TestExecuteEnqueuePayloadIsEnvelope(t *testing.T) {
	enqueuer := &recordingEnqueuer{}
	now := time.Date(2026, 3, 29, 10, 0, 0, 0, time.UTC)
	uc := &RunWorkflow{
		Enqueuer:      enqueuer,
		Clock:         fixedClock{now: now},
		ProducerTopic: "dagruntime-tasks",
	}

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
	if result.MessagingDestination != "dagruntime-tasks" {
		t.Fatalf("expected enqueue destination dagruntime-tasks, got %q", result.MessagingDestination)
	}
	if _, ok := enqueuer.last.Payload.([]events.PayloadEnvelope); !ok {
		t.Fatalf("expected envelope payload, got %T", enqueuer.last.Payload)
	}
	if !enqueuer.last.EventTime.Equal(now) {
		t.Fatalf("expected event time %s, got %s", now, enqueuer.last.EventTime)
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
	if result.MessagingDestination != defaultDriverDestinationName {
		t.Fatalf("expected driver destination %q, got %q", defaultDriverDestinationName, result.MessagingDestination)
	}
	if _, ok := recorder.last.Payload.(engine.InputMap); !ok {
		t.Fatalf("expected input map payload, got %T", recorder.last.Payload)
	}
}
