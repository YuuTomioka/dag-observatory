package engine

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	dagerrors "dag-observatory/dag-core/internal/domain/dagruntime/errors"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"
	"dag-observatory/dag-core/internal/domain/dagruntime/policy"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	artifactinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
)

type counterNode struct {
	mu    sync.Mutex
	count int
	spec  node.ExecutionSpec
	err   error
}

func (n *counterNode) Name() string                { return "counter.node" }
func (n *counterNode) Requires() []artifact.AnyKey { return nil }
func (n *counterNode) Provides() []artifact.AnyKey { return nil }
func (n *counterNode) Reads() []state.AnyKey       { return nil }
func (n *counterNode) Writes() []state.AnyKey      { return nil }
func (n *counterNode) Spec() node.ExecutionSpec    { return n.spec }
func (n *counterNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = av
	_ = aw
	_ = txn
	n.mu.Lock()
	n.count++
	n.mu.Unlock()
	return n.err
}

func (n *counterNode) Count() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.count
}

type timeoutNode struct {
	spec node.ExecutionSpec
}

func (n *timeoutNode) Name() string                { return "timeout.node" }
func (n *timeoutNode) Requires() []artifact.AnyKey { return nil }
func (n *timeoutNode) Provides() []artifact.AnyKey { return nil }
func (n *timeoutNode) Reads() []state.AnyKey       { return nil }
func (n *timeoutNode) Writes() []state.AnyKey      { return nil }
func (n *timeoutNode) Spec() node.ExecutionSpec    { return n.spec }
func (n *timeoutNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = av
	_ = aw
	_ = txn
	<-ctx.Done()
	return ctx.Err()
}

type sequenceObserver struct {
	mu      sync.Mutex
	events  []string
	starts  []port.NodeInfo
	results []port.NodeResult
}

type captureRecorder struct {
	mu         sync.Mutex
	nodeEvents []events.NodeExecutionEvent
}

func (r *captureRecorder) RecordEvent(ctx context.Context, event events.Event) {
	_ = ctx
	_ = event
}

func (r *captureRecorder) RecordNodeExecution(ctx context.Context, event events.NodeExecutionEvent) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodeEvents = append(r.nodeEvents, event)
}

func (r *captureRecorder) RecordNodeResult(ctx context.Context, result port.NodeResult) {
	_ = ctx
	_ = result
}

func (r *captureRecorder) RecordCycleResult(ctx context.Context, result port.CycleResult) {
	_ = ctx
	_ = result
}

func (r *captureRecorder) NodeEvents() []events.NodeExecutionEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]events.NodeExecutionEvent, len(r.nodeEvents))
	copy(out, r.nodeEvents)
	return out
}

func (o *sequenceObserver) OnCompile(ctx context.Context, info port.CompileInfo) {}
func (o *sequenceObserver) OnCycleStart(ctx context.Context, info port.CycleInfo) {
	o.append("cycle_start")
}
func (o *sequenceObserver) OnCycleEnd(ctx context.Context, info port.CycleResult) {
	o.append("cycle_end")
}
func (o *sequenceObserver) OnNodeStart(ctx context.Context, info port.NodeInfo) {
	o.append("node_start:" + info.NodeName)
	o.mu.Lock()
	o.starts = append(o.starts, info)
	o.mu.Unlock()
}
func (o *sequenceObserver) OnNodeEnd(ctx context.Context, info port.NodeResult) {
	o.append("node_end:" + info.NodeName)
	o.mu.Lock()
	o.results = append(o.results, info)
	o.mu.Unlock()
}
func (o *sequenceObserver) OnError(ctx context.Context, info port.ErrorInfo) {}

func (o *sequenceObserver) append(value string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.events = append(o.events, value)
}

func (o *sequenceObserver) Events() []string {
	o.mu.Lock()
	defer o.mu.Unlock()
	out := make([]string, len(o.events))
	copy(out, o.events)
	return out
}

func (o *sequenceObserver) Results() []port.NodeResult {
	o.mu.Lock()
	defer o.mu.Unlock()
	out := make([]port.NodeResult, len(o.results))
	copy(out, o.results)
	return out
}

func (o *sequenceObserver) Starts() []port.NodeInfo {
	o.mu.Lock()
	defer o.mu.Unlock()
	out := make([]port.NodeInfo, len(o.starts))
	copy(out, o.starts)
	return out
}

func TestRunnerRetryRespectsIdempotency(t *testing.T) {
	nodeNoRetry := &counterNode{
		spec: node.ExecutionSpec{SideEffect: true, Idempotent: false},
		err:  errors.New("fail"),
	}

	compiled := pipeline.Compiled{
		Name:  "retry",
		Order: []node.Node{nodeNoRetry},
		Nodes: []node.Node{nodeNoRetry},
	}

	runner := &Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{DefaultRetry: policy.RetryPolicy{MaxAttempts: 3}},
	}

	event := events.Event{
		EventID:   "retry",
		EventTime: time.Now(),
		Partition: state.Partition("default"),
		Type:      "test",
	}

	_ = runner.RunCycle(context.Background(), compiled, InputMap{}, event.Partition, event)

	if nodeNoRetry.Count() != 1 {
		t.Fatalf("expected no retry when not idempotent, got %d runs", nodeNoRetry.Count())
	}
}

func TestRunnerRetryWhenIdempotent(t *testing.T) {
	nodeRetry := &counterNode{
		spec: node.ExecutionSpec{SideEffect: true, Idempotent: true},
		err:  errors.New("fail"),
	}

	compiled := pipeline.Compiled{
		Name:  "retry",
		Order: []node.Node{nodeRetry},
		Nodes: []node.Node{nodeRetry},
	}

	runner := &Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{DefaultRetry: policy.RetryPolicy{MaxAttempts: 2}},
	}

	event := events.Event{
		EventID:   "retry",
		EventTime: time.Now(),
		Partition: state.Partition("default"),
		Type:      "test",
	}

	_ = runner.RunCycle(context.Background(), compiled, InputMap{}, event.Partition, event)

	if nodeRetry.Count() != 2 {
		t.Fatalf("expected retry for idempotent node, got %d runs", nodeRetry.Count())
	}
}

func TestRunnerRetryRespectsMaxAttemptsOne(t *testing.T) {
	nodeRetry := &counterNode{
		spec: node.ExecutionSpec{SideEffect: true, Idempotent: true},
		err:  errors.New("fail"),
	}

	compiled := pipeline.Compiled{
		Name:  "retry",
		Order: []node.Node{nodeRetry},
		Nodes: []node.Node{nodeRetry},
	}

	runner := &Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{DefaultRetry: policy.RetryPolicy{MaxAttempts: 1}},
	}

	event := events.Event{
		EventID:   "retry",
		EventTime: time.Now(),
		Partition: state.Partition("default"),
		Type:      "test",
	}

	_ = runner.RunCycle(context.Background(), compiled, InputMap{}, event.Partition, event)

	if nodeRetry.Count() != 1 {
		t.Fatalf("expected 1 attempt, got %d runs", nodeRetry.Count())
	}
}

func TestRunnerTimeout(t *testing.T) {
	timeout := &timeoutNode{
		spec: node.ExecutionSpec{Timeout: 5 * time.Millisecond},
	}

	compiled := pipeline.Compiled{
		Name:  "timeout",
		Order: []node.Node{timeout},
		Nodes: []node.Node{timeout},
	}

	runner := &Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{},
	}

	event := events.Event{
		EventID:   "timeout",
		EventTime: time.Now(),
		Partition: state.Partition("default"),
		Type:      "test",
	}

	err := runner.RunCycle(context.Background(), compiled, InputMap{}, event.Partition, event)
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

func TestRunnerDefaultTimeout(t *testing.T) {
	timeout := &timeoutNode{
		spec: node.ExecutionSpec{},
	}

	compiled := pipeline.Compiled{
		Name:  "timeout",
		Order: []node.Node{timeout},
		Nodes: []node.Node{timeout},
	}

	runner := &Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{DefaultTimeout: 5 * time.Millisecond},
	}

	event := events.Event{
		EventID:   "timeout",
		EventTime: time.Now(),
		Partition: state.Partition("default"),
		Type:      "test",
	}

	err := runner.RunCycle(context.Background(), compiled, InputMap{}, event.Partition, event)
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

func TestRunnerRecordsNodeExecutionEvents(t *testing.T) {
	nodeOK := &counterNode{
		spec: node.ExecutionSpec{Idempotent: true},
	}
	nodeFail := &counterNode{
		spec: node.ExecutionSpec{Idempotent: true},
		err:  errors.New("node failed"),
	}
	compiled := pipeline.Compiled{
		Name:  "record-node-execution",
		Order: []node.Node{nodeOK, nodeFail},
		Nodes: []node.Node{nodeOK, nodeFail},
	}
	recorder := &captureRecorder{}
	runner := &Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{DefaultRetry: policy.RetryPolicy{MaxAttempts: 1}},
		Recorder:      recorder,
	}
	eventTime := time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)
	event := events.Event{
		EventID:   "run-1",
		EventTime: eventTime,
		Partition: state.Partition("p1"),
		Type:      "test.requested",
	}

	_ = runner.RunCycle(context.Background(), compiled, InputMap{}, event.Partition, event)

	got := recorder.NodeEvents()
	if len(got) != 2 {
		t.Fatalf("expected 2 node execution events, got %d", len(got))
	}
	if got[0].RunID != "run-1" || got[1].RunID != "run-1" {
		t.Fatalf("expected run_id run-1, got %#v", got)
	}
	if got[0].SequenceNo != 1 || got[1].SequenceNo != 2 {
		t.Fatalf("expected sequence numbers [1,2], got [%d,%d]", got[0].SequenceNo, got[1].SequenceNo)
	}
	if got[0].Status != events.NodeExecutionStatusSucceeded {
		t.Fatalf("expected first status succeeded, got %s", got[0].Status)
	}
	if got[1].Status != events.NodeExecutionStatusFailed {
		t.Fatalf("expected second status failed, got %s", got[1].Status)
	}
	if got[0].TriggerReason != "event_dispatch" || got[1].TriggerReason != "event_dispatch" {
		t.Fatalf("expected trigger_reason event_dispatch, got %#v", got)
	}
	if got[0].Error != "" {
		t.Fatalf("expected no error on first event, got %q", got[0].Error)
	}
	if got[1].Error == "" {
		t.Fatal("expected error text on failed event")
	}
	if !got[0].EventTime.Equal(eventTime) || !got[1].EventTime.Equal(eventTime) {
		t.Fatalf("expected event time %s, got %#v", eventTime, got)
	}
	if got[0].NodeName == "" || got[1].NodeName == "" {
		t.Fatalf("expected node names, got %#v", got)
	}
	if got[0].DurationNS <= 0 || got[1].DurationNS <= 0 {
		t.Fatalf("expected positive duration_ns, got %#v", got)
	}
}

type skipNode struct {
	spec   node.ExecutionSpec
	reason string
}

func (n *skipNode) Name() string                { return "skip.node" }
func (n *skipNode) Requires() []artifact.AnyKey { return nil }
func (n *skipNode) Provides() []artifact.AnyKey { return nil }
func (n *skipNode) Reads() []state.AnyKey       { return nil }
func (n *skipNode) Writes() []state.AnyKey      { return nil }
func (n *skipNode) Spec() node.ExecutionSpec    { return n.spec }
func (n *skipNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = av
	_ = aw
	_ = txn
	return dagerrors.SkipError{Reason: n.reason}
}

func TestRunnerSkipErrorIsRecordedAndDoesNotFailCycle(t *testing.T) {
	skipped := &skipNode{
		spec:   node.ExecutionSpec{Idempotent: true},
		reason: "session blocked",
	}
	next := &counterNode{
		spec: node.ExecutionSpec{Idempotent: true},
	}
	compiled := pipeline.Compiled{
		Name:  "skip-node",
		Order: []node.Node{skipped, next},
		Nodes: []node.Node{skipped, next},
	}
	recorder := &captureRecorder{}
	runner := &Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{DefaultRetry: policy.RetryPolicy{MaxAttempts: 1}},
		Recorder:      recorder,
	}
	event := events.Event{
		EventID:   "run-skip",
		EventTime: time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC),
		Partition: state.Partition("p1"),
		Type:      "test.requested",
	}

	err := runner.RunCycle(context.Background(), compiled, InputMap{}, event.Partition, event)
	if err != nil {
		t.Fatalf("expected skip to not fail cycle, got %v", err)
	}
	if next.Count() != 1 {
		t.Fatalf("expected next node to run, got %d", next.Count())
	}
	got := recorder.NodeEvents()
	if len(got) != 2 {
		t.Fatalf("expected 2 node execution events, got %d", len(got))
	}
	if got[0].Status != events.NodeExecutionStatusSkipped {
		t.Fatalf("expected first status skipped, got %s", got[0].Status)
	}
	if got[0].SkipReason != "session_blocked" {
		t.Fatalf("expected normalized skip reason session_blocked, got %q", got[0].SkipReason)
	}
	if got[0].Error != "" {
		t.Fatalf("expected empty error for skipped node, got %q", got[0].Error)
	}
}

func TestRunnerTriggerReasonRetry(t *testing.T) {
	nodeRetry := &counterNode{
		spec: node.ExecutionSpec{SideEffect: true, Idempotent: true},
		err:  errors.New("retry target"),
	}
	compiled := pipeline.Compiled{
		Name:  "retry-trigger-reason",
		Order: []node.Node{nodeRetry},
		Nodes: []node.Node{nodeRetry},
	}
	recorder := &captureRecorder{}
	runner := &Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{DefaultRetry: policy.RetryPolicy{MaxAttempts: 2}},
		Recorder:      recorder,
	}
	event := events.Event{
		EventID:   "run-retry",
		EventTime: time.Date(2026, 4, 11, 12, 30, 0, 0, time.UTC),
		Partition: state.Partition("p1"),
		Type:      "test.requested",
	}

	_ = runner.RunCycle(context.Background(), compiled, InputMap{}, event.Partition, event)

	got := recorder.NodeEvents()
	if len(got) != 2 {
		t.Fatalf("expected 2 node execution events for retry, got %d", len(got))
	}
	if got[0].TriggerReason != "event_dispatch" {
		t.Fatalf("expected first trigger reason event_dispatch, got %q", got[0].TriggerReason)
	}
	if got[1].TriggerReason != "retry" {
		t.Fatalf("expected second trigger reason retry, got %q", got[1].TriggerReason)
	}
}

func TestRunnerTimeoutWrapsTimeoutError(t *testing.T) {
	timeout := &timeoutNode{
		spec: node.ExecutionSpec{Timeout: 5 * time.Millisecond},
	}

	compiled := pipeline.Compiled{
		Name:  "timeout",
		Order: []node.Node{timeout},
		Nodes: []node.Node{timeout},
	}

	runner := &Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{DefaultRetry: policy.RetryPolicy{MaxAttempts: 1}},
	}

	event := events.Event{
		EventID:   "timeout",
		EventTime: time.Now(),
		Partition: state.Partition("default"),
		Type:      "test",
	}

	err := runner.RunCycle(context.Background(), compiled, InputMap{}, event.Partition, event)
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	var timeoutErr dagerrors.TimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("expected TimeoutError, got %T", err)
	}
	if timeoutErr.Node != "timeout.node" {
		t.Fatalf("expected node name, got %q", timeoutErr.Node)
	}
}

func TestObserverSequence(t *testing.T) {
	n1 := &counterNode{}
	n2 := &counterNode{}

	compiled := pipeline.Compiled{
		Name:  "obs",
		Order: []node.Node{n1, n2},
		Nodes: []node.Node{n1, n2},
	}

	observer := &sequenceObserver{}

	runner := &Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{},
		Observer:      observer,
	}

	event := events.Event{
		EventID:   "obs",
		EventTime: time.Now(),
		Partition: state.Partition("default"),
		Type:      "test",
	}

	if err := runner.RunCycle(context.Background(), compiled, InputMap{}, event.Partition, event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := observer.Events()
	if len(events) != 6 {
		t.Fatalf("expected 6 observer events, got %d: %v", len(events), events)
	}
	if events[0] != "cycle_start" || events[5] != "cycle_end" {
		t.Fatalf("unexpected cycle boundaries: %v", events)
	}
	if events[1] != "node_start:counter.node" || events[2] != "node_end:counter.node" {
		t.Fatalf("unexpected node 1 events: %v", events[1:3])
	}
	if events[3] != "node_start:counter.node" || events[4] != "node_end:counter.node" {
		t.Fatalf("unexpected node 2 events: %v", events[3:5])
	}

	results := observer.Results()
	for _, res := range results {
		if res.RetryCount != 0 {
			t.Fatalf("expected retry count 0, got %d", res.RetryCount)
		}
		if res.QueueWaitMS < 0 {
			t.Fatalf("expected non-negative queue wait, got %d", res.QueueWaitMS)
		}
	}
}

func TestQueueWaitObservedFromEventTime(t *testing.T) {
	n1 := &counterNode{}
	compiled := pipeline.Compiled{
		Name:  "queue_wait",
		Order: []node.Node{n1},
		Nodes: []node.Node{n1},
	}
	observer := &sequenceObserver{}
	runner := &Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{},
		Observer:      observer,
	}
	event := events.Event{
		EventID:   "queue_wait",
		EventTime: time.Now().Add(-2 * time.Second),
		Partition: state.Partition("default"),
		Type:      "test",
	}
	if err := runner.RunCycle(context.Background(), compiled, InputMap{}, event.Partition, event); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	starts := observer.Starts()
	if len(starts) != 1 {
		t.Fatalf("expected one node start, got %d", len(starts))
	}
	if starts[0].QueueWaitMS <= 0 {
		t.Fatalf("expected positive queue wait, got %d", starts[0].QueueWaitMS)
	}
	results := observer.Results()
	if len(results) != 1 {
		t.Fatalf("expected one node result, got %d", len(results))
	}
	if results[0].QueueWaitMS <= 0 {
		t.Fatalf("expected positive queue wait in result, got %d", results[0].QueueWaitMS)
	}
}
