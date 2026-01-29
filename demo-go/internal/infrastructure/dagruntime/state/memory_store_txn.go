package state

import (
	"sync"

	domain "dag-observatory/demo-go/internal/domain/dagruntime/state"
)

type MemoryStore struct {
	mu   sync.RWMutex
	data map[domain.Partition]map[domain.RawKey]any
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: map[domain.Partition]map[domain.RawKey]any{},
	}
}

func (s *MemoryStore) BeginTxn(part domain.Partition) domain.Txn {
	return &MemoryTxn{
		store:     s,
		partition: part,
		staged:    map[domain.RawKey]any{},
	}
}

type MemoryTxn struct {
	store     *MemoryStore
	partition domain.Partition
	staged    map[domain.RawKey]any
	closed    bool
}

func (t *MemoryTxn) Get(key domain.AnyKey) (any, bool) {
	raw := key.Raw()
	if value, ok := t.staged[raw]; ok {
		return value, true
	}
	t.store.mu.RLock()
	defer t.store.mu.RUnlock()
	partData := t.store.data[t.partition]
	if partData == nil {
		return nil, false
	}
	value, ok := partData[raw]
	return value, ok
}

func (t *MemoryTxn) StageWrite(key domain.AnyKey, value any) {
	if t.closed {
		return
	}
	t.staged[key.Raw()] = value
}

func (t *MemoryTxn) Commit() error {
	if t.closed {
		return nil
	}
	t.store.mu.Lock()
	defer t.store.mu.Unlock()
	partData := t.store.data[t.partition]
	if partData == nil {
		partData = map[domain.RawKey]any{}
		t.store.data[t.partition] = partData
	}
	for key, value := range t.staged {
		partData[key] = value
	}
	t.closed = true
	return nil
}

func (t *MemoryTxn) Rollback() {
	t.staged = map[domain.RawKey]any{}
	t.closed = true
}

