package events

import "testing"

type samplePayload struct {
	Name string `json:"name"`
}

func TestPayloadCodecRoundTrip(t *testing.T) {
	key := PayloadKey[samplePayload]{
		Name:     "sample",
		StableID: "event:sample.v1",
	}
	orig := samplePayload{Name: "ok"}

	env, err := EncodePayload(key, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	decoded, err := DecodePayload(key, env)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.Name != orig.Name {
		t.Fatalf("expected %q, got %q", orig.Name, decoded.Name)
	}
}

func TestPayloadCodecKeyMismatch(t *testing.T) {
	key := PayloadKey[samplePayload]{
		Name:     "sample",
		StableID: "event:sample.v1",
	}
	other := PayloadKey[samplePayload]{
		Name:     "other",
		StableID: "event:other.v1",
	}
	env, err := EncodePayload(other, samplePayload{Name: "ng"})
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	if _, err := DecodePayload(key, env); err == nil {
		t.Fatalf("expected key mismatch error")
	}
}
