package events

import (
	"encoding/json"
	"fmt"
)

type PayloadKey[T any] struct {
	Name     string
	StableID string
	Codec    Codec[T]
}

type RawPayloadKey struct {
	Name     string
	StableID string
}

type AnyPayloadKey interface {
	Raw() RawPayloadKey
	String() string
}

func (k PayloadKey[T]) Raw() RawPayloadKey {
	return RawPayloadKey{Name: k.Name, StableID: k.StableID}
}

func (k PayloadKey[T]) String() string {
	if k.StableID == "" {
		return k.Name
	}
	if k.Name == "" {
		return k.StableID
	}
	return fmt.Sprintf("%s:%s", k.StableID, k.Name)
}

type Codec[T any] interface {
	Encode(value T) ([]byte, error)
	Decode(data []byte) (T, error)
}

type JSONCodec[T any] struct{}

func (c JSONCodec[T]) Encode(value T) ([]byte, error) {
	return json.Marshal(value)
}

func (c JSONCodec[T]) Decode(data []byte) (T, error) {
	var out T
	err := json.Unmarshal(data, &out)
	return out, err
}

type PayloadEnvelope struct {
	Key  RawPayloadKey
	Data []byte
}

func EncodePayload[T any](key PayloadKey[T], value T) (PayloadEnvelope, error) {
	codec := key.Codec
	if codec == nil {
		codec = JSONCodec[T]{}
	}
	data, err := codec.Encode(value)
	if err != nil {
		return PayloadEnvelope{}, err
	}
	return PayloadEnvelope{
		Key:  key.Raw(),
		Data: data,
	}, nil
}

func DecodePayload[T any](key PayloadKey[T], payload PayloadEnvelope) (T, error) {
	if payload.Key != key.Raw() {
		var zero T
		return zero, fmt.Errorf("payload key mismatch: expected %s", key.String())
	}
	codec := key.Codec
	if codec == nil {
		codec = JSONCodec[T]{}
	}
	return codec.Decode(payload.Data)
}
