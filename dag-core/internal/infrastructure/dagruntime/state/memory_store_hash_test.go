package state

import (
	"testing"

	domain "dag-observatory/dag-core/internal/domain/dagruntime/state"
)

func TestMemoryStoreHashStable(t *testing.T) {
	store := NewMemoryStore()
	part := domain.Partition("USDJPY")
	key := domain.Key[string]{Name: "mode", StableID: "state:test.mode.v1"}

	txn := store.BeginTxn(part)
	domain.StageWrite(txn, key, "debug")
	if err := txn.Commit(); err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	h1 := store.Hash(part)
	h2 := store.Hash(part)
	if h1 == "" || h2 == "" {
		t.Fatalf("expected non-empty hash")
	}
	if h1 != h2 {
		t.Fatalf("expected stable hash, got %s vs %s", h1, h2)
	}
}
