package compiler

import (
	"path/filepath"
	"strings"
	"testing"

	"dag-observatory/dag-core/internal/application/dagruntime/spec/factory"
)

func TestCompileFromPath(t *testing.T) {
	t.Parallel()

	registry, err := factory.NewBuiltinRegistry()
	if err != nil {
		t.Fatalf("new builtin registry: %v", err)
	}
	c := NewDefault()
	c.ResolveSpecPath = func(path string) (string, error) {
		return filepath.Join("..", "..", "..", "..", "..", path), nil
	}
	compiled, err := c.CompileFromPath("workflows/default.yaml", registry)
	if err != nil {
		t.Fatalf("compile from path: %v", err)
	}
	if compiled.Name != "dagruntime.default" {
		t.Fatalf("expected workflow name dagruntime.default, got %q", compiled.Name)
	}
}

func TestCompileFromPathNilRegistry(t *testing.T) {
	t.Parallel()

	_, err := NewDefault().CompileFromPath("workflows/default.yaml", nil)
	if err == nil {
		t.Fatal("expected error for nil registry")
	}
}

func TestResolveSpecPathNotFound(t *testing.T) {
	t.Parallel()

	_, err := ResolveSpecPath("workflows/does_not_exist.yaml")
	if err == nil {
		t.Fatal("expected not found error")
	}
	if !strings.Contains(err.Error(), "workflow spec not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
