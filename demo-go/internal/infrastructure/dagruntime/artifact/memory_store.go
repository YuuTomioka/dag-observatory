package artifact

import (
	"sync"

	domain "dag-observatory/demo-go/internal/domain/dagruntime/artifact"
)

type MemoryStore struct {
	mu   sync.RWMutex
	data map[domain.RawKey]any
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: map[domain.RawKey]any{},
	}
}

func (s *MemoryStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = map[domain.RawKey]any{}
}

func (s *MemoryStore) Get(key domain.AnyKey) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.data[key.Raw()]
	return value, ok
}

func (s *MemoryStore) Set(key domain.AnyKey, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key.Raw()] = value
}

func (s *MemoryStore) MustGet(key domain.AnyKey) any {
	value, ok := s.Get(key)
	if !ok {
		panic("artifact: missing key " + key.String())
	}
	return value
}

func (s *MemoryStore) Has(key domain.AnyKey) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key.Raw()]
	return ok
}

func (s *MemoryStore) View() domain.View {
	return s
}

