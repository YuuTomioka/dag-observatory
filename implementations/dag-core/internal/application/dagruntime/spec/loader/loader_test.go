package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.yaml")
	content := `name: dagruntime.default
version: v1
inputs:
  - symbol
  - mode
nodes:
  - id: heavy_calc
    kind: heavy_calc
    config: {}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write workflow yaml: %v", err)
	}

	l := Loader{}
	spec, err := l.LoadFile(path)
	if err != nil {
		t.Fatalf("load file: %v", err)
	}
	if spec.Name != "dagruntime.default" {
		t.Fatalf("expected name dagruntime.default, got %q", spec.Name)
	}
	if len(spec.Nodes) != 1 {
		t.Fatalf("expected one node, got %d", len(spec.Nodes))
	}
}

func TestLoadFileDecodeError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.yaml")
	content := `name: dagruntime.default
version: v1
nodes:
  - id: heavy_calc
    kind: heavy_calc
    config: [unclosed
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write workflow yaml: %v", err)
	}

	l := Loader{}
	if _, err := l.LoadFile(path); err == nil {
		t.Fatal("expected decode error")
	}
}
