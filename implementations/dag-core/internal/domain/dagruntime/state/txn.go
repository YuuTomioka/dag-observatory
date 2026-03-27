package state

import "fmt"

type Txn interface {
	Get(key AnyKey) (any, bool)
	StageWrite(key AnyKey, value any)
	Commit() error
	Rollback()
}

func Get[T any](txn Txn, key Key[T]) (T, bool) {
	value, ok := txn.Get(key)
	if !ok {
		var zero T
		return zero, false
	}
	typed, ok := value.(T)
	if !ok {
		var zero T
		return zero, false
	}
	return typed, true
}

func MustGet[T any](txn Txn, key Key[T]) T {
	value, ok := txn.Get(key)
	if !ok {
		panic(fmt.Sprintf("state: missing key %s", key.String()))
	}
	typed, ok := value.(T)
	if !ok {
		panic(fmt.Sprintf("state: type mismatch for key %s", key.String()))
	}
	return typed
}

func StageWrite[T any](txn Txn, key Key[T], value T) {
	txn.StageWrite(key, value)
}

