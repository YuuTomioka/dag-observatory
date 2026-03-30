package engine

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"
	"dag-observatory/dag-core/internal/domain/dagruntime/policy"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	artifactinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
)

type writeStateNode struct {
	key   state.Key[int]
	value int
}

func (n *writeStateNode) Name() string                { return "write.state" }
func (n *writeStateNode) Requires() []artifact.AnyKey { return nil }
func (n *writeStateNode) Provides() []artifact.AnyKey { return nil }
func (n *writeStateNode) Reads() []state.AnyKey       { return nil }
func (n *writeStateNode) Writes() []state.AnyKey      { return []state.AnyKey{n.key} }
func (n *writeStateNode) Spec() node.ExecutionSpec    { return node.ExecutionSpec{} }
func (n *writeStateNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = av
	_ = aw
	state.StageWrite(txn, n.key, n.value)
	return nil
}

type readStateNode struct {
	key   state.Key[int]
	value int
}

func (n *readStateNode) Name() string                { return "read.state" }
func (n *readStateNode) Requires() []artifact.AnyKey { return nil }
func (n *readStateNode) Provides() []artifact.AnyKey { return nil }
func (n *readStateNode) Reads() []state.AnyKey       { return []state.AnyKey{n.key} }
func (n *readStateNode) Writes() []state.AnyKey      { return nil }
func (n *readStateNode) Spec() node.ExecutionSpec    { return node.ExecutionSpec{} }
func (n *readStateNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = av
	_ = aw
	got, ok := state.Get(txn, n.key)
	if !ok || got != n.value {
		return errExpectedState
	}
	return nil
}

var errExpectedState = stateError("expected state value not found")

type stateError string

func (e stateError) Error() string { return string(e) }

func TestBoltStorePersistenceE2E(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.db")
	part := state.Partition("USDJPY")
	key := state.Key[int]{Name: "count", StableID: "state:test.count.v1"}

	store, err := stateinfra.NewBoltStore(path, nil)
	if err != nil {
		t.Fatalf("open store failed: %v", err)
	}

	write := &writeStateNode{key: key, value: 7}
	compiled := pipeline.Compiled{
		Name:  "persist",
		Order: []node.Node{write},
		Nodes: []node.Node{write},
	}

	runner := &Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    store,
		Policy:        policy.Policy{DefaultRetry: policy.RetryPolicy{MaxAttempts: 1}},
	}

	event := events.Event{
		EventID:   "persist",
		EventTime: time.Now(),
		Partition: part,
		Type:      "test",
	}

	if err := runner.RunCycle(context.Background(), compiled, InputMap{}, part, event); err != nil {
		t.Fatalf("run cycle failed: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close store failed: %v", err)
	}

	store2, err := stateinfra.NewBoltStore(path, nil)
	if err != nil {
		t.Fatalf("reopen store failed: %v", err)
	}
	defer store2.Close()

	read := &readStateNode{key: key, value: 7}
	compiled2 := pipeline.Compiled{
		Name:  "persist",
		Order: []node.Node{read},
		Nodes: []node.Node{read},
	}

	runner2 := &Runner{
		ArtifactStore: artifactinfra.NewMemoryStore(),
		StateStore:    store2,
		Policy:        policy.Policy{DefaultRetry: policy.RetryPolicy{MaxAttempts: 1}},
	}

	if err := runner2.RunCycle(context.Background(), compiled2, InputMap{}, part, event); err != nil {
		t.Fatalf("run cycle failed after reopen: %v", err)
	}
}
