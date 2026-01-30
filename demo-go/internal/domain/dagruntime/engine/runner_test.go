package engine

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"dag-observatory/demo-go/internal/application/dagruntime/port"
	"dag-observatory/demo-go/internal/domain/dagruntime/artifact"
	"dag-observatory/demo-go/internal/domain/dagruntime/events"
	"dag-observatory/demo-go/internal/domain/dagruntime/node"
	"dag-observatory/demo-go/internal/domain/dagruntime/pipeline"
	"dag-observatory/demo-go/internal/domain/dagruntime/policy"
	"dag-observatory/demo-go/internal/domain/dagruntime/state"
	artifactinfra "dag-observatory/demo-go/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/demo-go/internal/infrastructure/dagruntime/state"
)

type counterNode struct {
	mu    sync.Mutex
	count int
	spec  node.ExecutionSpec
	err   error
}

func (n *counterNode) Requires() []artifact.AnyKey { return nil }
func (n *counterNode) Provides() []artifact.AnyKey { return nil }
func (n *counterNode) Reads() []state.AnyKey       { return nil }
func (n *counterNode) Writes() []state.AnyKey      { return nil }
func (n *counterNode) Spec() node.ExecutionSpec    { return n.spec }
func (n *counterNode) Run(ctx context.Context, av artifact.View, txn state.Txn) error {
	_ = ctx
	_ = av
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

func (n *timeoutNode) Requires() []artifact.AnyKey { return nil }
func (n *timeoutNode) Provides() []artifact.AnyKey { return nil }
func (n *timeoutNode) Reads() []state.AnyKey       { return nil }
func (n *timeoutNode) Writes() []state.AnyKey      { return nil }
func (n *timeoutNode) Spec() node.ExecutionSpec    { return n.spec }
func (n *timeoutNode) Run(ctx context.Context, av artifact.View, txn state.Txn) error {
	_ = av
	_ = txn
	<-ctx.Done()
	return ctx.Err()
}

type sequenceObserver struct {
	mu     sync.Mutex
	events []string
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
}
func (o *sequenceObserver) OnNodeEnd(ctx context.Context, info port.NodeResult) {
	o.append("node_end:" + info.NodeName)
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
	if events[1] != "node_start:*engine.counterNode" || events[2] != "node_end:*engine.counterNode" {
		t.Fatalf("unexpected node 1 events: %v", events[1:3])
	}
	if events[3] != "node_start:*engine.counterNode" || events[4] != "node_end:*engine.counterNode" {
		t.Fatalf("unexpected node 2 events: %v", events[3:5])
	}
}
