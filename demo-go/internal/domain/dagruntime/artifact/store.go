package artifact

import "fmt"

type View interface {
	Get(key AnyKey) (any, bool)
	MustGet(key AnyKey) any
	Has(key AnyKey) bool
}

type Store interface {
	Clear()
	Get(key AnyKey) (any, bool)
	Set(key AnyKey, value any)
	MustGet(key AnyKey) any
	Has(key AnyKey) bool
	View() View
}

func Get[T any](view View, key Key[T]) (T, bool) {
	value, ok := view.Get(key)
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

func MustGet[T any](view View, key Key[T]) T {
	value, ok := view.Get(key)
	if !ok {
		panic(fmt.Sprintf("artifact: missing key %s", key.String()))
	}
	typed, ok := value.(T)
	if !ok {
		panic(fmt.Sprintf("artifact: type mismatch for key %s", key.String()))
	}
	return typed
}

func Set[T any](store Store, key Key[T], value T) {
	store.Set(key, value)
}

