package driver

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/engine"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"
	"dag-observatory/dag-core/internal/domain/dagruntime/policy"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	artifactinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
)

type eventNode struct {
	mu          sync.Mutex
	successRuns int
	failRuns    int
}

type panicNode struct{}

func (n *panicNode) Name() string                { return "panic.node" }
func (n *panicNode) Requires() []artifact.AnyKey { return nil }
func (n *panicNode) Provides() []artifact.AnyKey { return nil }
func (n *panicNode) Reads() []state.AnyKey       { return nil }
func (n *panicNode) Writes() []state.AnyKey      { return nil }
func (n *panicNode) Spec() node.ExecutionSpec    { return node.ExecutionSpec{} }
func (n *panicNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = av
	_ = aw
	_ = txn
	panic("boom")
}

func (n *eventNode) Name() string                { return "event.node" }
func (n *eventNode) Requires() []artifact.AnyKey { return nil }
func (n *eventNode) Provides() []artifact.AnyKey { return nil }
func (n *eventNode) Reads() []state.AnyKey       { return nil }
func (n *eventNode) Writes() []state.AnyKey      { return nil }
func (n *eventNode) Spec() node.ExecutionSpec    { return node.ExecutionSpec{} }
func (n *eventNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = av
	_ = aw
	_ = txn
	ev, ok := events.EventFromContext(ctx)
	if ok && ev.EventID == "fail" {
		n.mu.Lock()
		n.failRuns++
		n.mu.Unlock()
		return errors.New("fail")
	}
	n.mu.Lock()
	n.successRuns++
	n.mu.Unlock()
	return nil
}

func (n *eventNode) Counts() (int, int) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.successRuns, n.failRuns
}

func TestDriverContinueOnError(t *testing.T) {
	n := &eventNode{}
	compiled := pipeline.Compiled{
		Name:  "driver",
		Order: []node.Node{n},
		Nodes: []node.Node{n},
	}
	runner := &engine.Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{DefaultRetry: policy.RetryPolicy{MaxAttempts: 1}},
	}
	d := &Driver{
		Runner:   runner,
		Compiled: compiled,
		Options:  engine.DriverOptions{ContinueOnError: true},
	}

	stream := make(chan events.Event, 2)
	stream <- events.Event{EventID: "fail", EventTime: time.Now()}
	stream <- events.Event{EventID: "ok", EventTime: time.Now()}
	close(stream)

	if err := d.Run(context.Background(), stream); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	success, fail := n.Counts()
	if success != 1 || fail != 1 {
		t.Fatalf("unexpected counts: success=%d fail=%d", success, fail)
	}
}

func TestDriverStopsOnError(t *testing.T) {
	n := &eventNode{}
	compiled := pipeline.Compiled{
		Name:  "driver",
		Order: []node.Node{n},
		Nodes: []node.Node{n},
	}
	runner := &engine.Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{DefaultRetry: policy.RetryPolicy{MaxAttempts: 1}},
	}
	d := &Driver{
		Runner:   runner,
		Compiled: compiled,
		Options:  engine.DriverOptions{ContinueOnError: false},
	}

	stream := make(chan events.Event, 2)
	stream <- events.Event{EventID: "fail", EventTime: time.Now()}
	stream <- events.Event{EventID: "ok", EventTime: time.Now()}
	close(stream)

	if err := d.Run(context.Background(), stream); err == nil {
		t.Fatalf("expected error")
	}

	success, fail := n.Counts()
	if success != 0 || fail != 1 {
		t.Fatalf("unexpected counts: success=%d fail=%d", success, fail)
	}
}

func TestDriverConvertsNodePanicToError(t *testing.T) {
	n := &panicNode{}
	compiled := pipeline.Compiled{
		Name:  "driver-panic",
		Order: []node.Node{n},
		Nodes: []node.Node{n},
	}
	runner := &engine.Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    stateinfra.NewMemoryStore(),
		Policy:        policy.Policy{DefaultRetry: policy.RetryPolicy{MaxAttempts: 1}},
	}
	d := &Driver{
		Runner:   runner,
		Compiled: compiled,
	}

	stream := make(chan events.Event, 1)
	stream <- events.Event{EventID: "panic-event", EventTime: time.Now()}
	close(stream)

	err := d.Run(context.Background(), stream)
	if err == nil {
		t.Fatal("expected error for panic node")
	}
	if got := err.Error(); got == "" || !strings.Contains(got, "panic-event") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
