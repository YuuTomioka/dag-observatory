package grpc

import "testing"

func TestNewServerSmoke(t *testing.T) {
	srv := New(Dependencies{})
	if srv == nil {
		t.Fatal("expected server to be created")
	}
}
