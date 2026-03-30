package state

import "fmt"

type RawKey struct {
	Name     string
	StableID string
}

type AnyKey interface {
	Raw() RawKey
	String() string
}

type Key[T any] struct {
	Name     string
	StableID string
}

func (k Key[T]) Raw() RawKey {
	return RawKey{Name: k.Name, StableID: k.StableID}
}

func (k Key[T]) String() string {
	if k.StableID == "" {
		return k.Name
	}
	if k.Name == "" {
		return k.StableID
	}
	return fmt.Sprintf("%s:%s", k.StableID, k.Name)
}
