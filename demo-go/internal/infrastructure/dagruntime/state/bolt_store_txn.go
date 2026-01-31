package state

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"

	domain "dag-observatory/demo-go/internal/domain/dagruntime/state"
	"go.etcd.io/bbolt"
)

const boltBucketName = "dagruntime_state"

type BoltStore struct {
	db *bbolt.DB
	mu sync.Mutex
}

func NewBoltStore(path string, opts *bbolt.Options) (*BoltStore, error) {
	db, err := bbolt.Open(path, 0o600, opts)
	if err != nil {
		return nil, err
	}
	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(boltBucketName))
		return err
	}); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &BoltStore{db: db}, nil
}

func (s *BoltStore) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *BoltStore) BeginTxn(part domain.Partition) domain.Txn {
	return &BoltTxn{
		store:     s,
		partition: part,
		staged:    map[string][]byte{},
	}
}

func (s *BoltStore) Hash(part domain.Partition) string {
	if s.db == nil {
		return ""
	}
	var entries []kvEntry
	_ = s.db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte(boltBucketName))
		if root == nil {
			return nil
		}
		b := root.Bucket([]byte(part))
		if b == nil {
			return nil
		}
		_ = b.ForEach(func(k, v []byte) error {
			entries = append(entries, kvEntry{key: string(k), value: v})
			return nil
		})
		return nil
	})
	if len(entries) == 0 {
		return ""
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})
	hasher := sha256.New()
	for _, entry := range entries {
		_, _ = fmt.Fprintf(hasher, "%s|", entry.key)
		hasher.Write(entry.value)
		hasher.Write([]byte{'\n'})
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

type BoltTxn struct {
	store     *BoltStore
	partition domain.Partition
	staged    map[string][]byte
	closed    bool
}

func (t *BoltTxn) Get(key domain.AnyKey) (any, bool) {
	raw := encodeKey(key.Raw())
	if value, ok := t.staged[raw]; ok {
		return decodeValue(value)
	}
	if t.store == nil || t.store.db == nil {
		return nil, false
	}
	var data []byte
	_ = t.store.db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte(boltBucketName))
		if root == nil {
			return nil
		}
		b := root.Bucket([]byte(t.partition))
		if b == nil {
			return nil
		}
		if value := b.Get([]byte(raw)); value != nil {
			data = append([]byte(nil), value...)
		}
		return nil
	})
	if data == nil {
		return nil, false
	}
	return decodeValue(data)
}

func (t *BoltTxn) StageWrite(key domain.AnyKey, value any) {
	if t.closed {
		return
	}
	data, err := encodeValue(value)
	if err != nil {
		return
	}
	t.staged[encodeKey(key.Raw())] = data
}

func (t *BoltTxn) Commit() error {
	if t.closed || t.store == nil || t.store.db == nil {
		return nil
	}
	t.store.mu.Lock()
	defer t.store.mu.Unlock()
	err := t.store.db.Update(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte(boltBucketName))
		if root == nil {
			return fmt.Errorf("bolt: missing root bucket")
		}
		b, err := root.CreateBucketIfNotExists([]byte(t.partition))
		if err != nil {
			return err
		}
		for key, value := range t.staged {
			if err := b.Put([]byte(key), value); err != nil {
				return err
			}
		}
		return nil
	})
	t.closed = true
	return err
}

func (t *BoltTxn) Rollback() {
	t.staged = map[string][]byte{}
	t.closed = true
}

type kvEntry struct {
	key   string
	value []byte
}

func encodeKey(key domain.RawKey) string {
	if key.StableID == "" {
		return key.Name
	}
	if key.Name == "" {
		return key.StableID
	}
	return key.StableID + "|" + key.Name
}

func encodeValue(value any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(storedValue{Value: value}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeValue(data []byte) (any, bool) {
	if len(data) == 0 {
		return nil, false
	}
	var wrapper storedValue
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&wrapper); err != nil {
		return nil, false
	}
	return wrapper.Value, true
}

type storedValue struct {
	Value any
}
