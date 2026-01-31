package state

import (
	"path/filepath"
	"testing"

	domain "dag-observatory/demo-go/internal/domain/dagruntime/state"
)

func TestBoltStoreHashStableAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.db")
	store, err := NewBoltStore(path, nil)
	if err != nil {
		t.Fatalf("open store failed: %v", err)
	}

	part := domain.Partition("USDJPY")
	key := domain.Key[string]{Name: "mode", StableID: "state:test.mode.v1"}
	txn := store.BeginTxn(part)
	domain.StageWrite(txn, key, "debug")
	if err := txn.Commit(); err != nil {
		t.Fatalf("commit failed: %v", err)
	}
	h1 := store.Hash(part)
	if err := store.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	store2, err := NewBoltStore(path, nil)
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	defer store2.Close()
	h2 := store2.Hash(part)

	if h1 == "" || h2 == "" {
		t.Fatalf("expected non-empty hash")
	}
	if h1 != h2 {
		t.Fatalf("expected stable hash, got %s vs %s", h1, h2)
	}
}
