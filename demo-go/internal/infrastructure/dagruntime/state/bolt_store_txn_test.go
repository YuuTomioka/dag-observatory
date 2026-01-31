package state

import (
	"encoding/gob"
	"path/filepath"
	"testing"

	domain "dag-observatory/demo-go/internal/domain/dagruntime/state"
)

func TestBoltStorePersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.db")
	gob.Register(int(0))

	store, err := NewBoltStore(path, nil)
	if err != nil {
		t.Fatalf("open store failed: %v", err)
	}

	part := domain.Partition("USDJPY")
	key := domain.Key[int]{Name: "count", StableID: "state:test.count.v1"}

	txn := store.BeginTxn(part)
	domain.StageWrite(txn, key, 1)
	if err := txn.Commit(); err != nil {
		t.Fatalf("commit failed: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	store2, err := NewBoltStore(path, nil)
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	defer store2.Close()

	txn2 := store2.BeginTxn(part)
	value, ok := domain.Get(txn2, key)
	if !ok {
		t.Fatalf("expected value after reopen")
	}
	if value != 1 {
		t.Fatalf("expected 1, got %d", value)
	}
	txn2.Rollback()
}
