package di

import "testing"

func TestCompileDefaultWorkflowFromYAML(t *testing.T) {
	t.Parallel()

	compiled, err := compileDefaultWorkflow(Config{WorkflowSpecPath: "workflows/default.yaml"})
	if err != nil {
		t.Fatalf("compile default workflow: %v", err)
	}
	if compiled.Name != "dagruntime.default" {
		t.Fatalf("expected workflow name dagruntime.default, got %q", compiled.Name)
	}
	if len(compiled.Nodes) != 1 {
		t.Fatalf("expected one node, got %d", len(compiled.Nodes))
	}
	if len(compiled.Inputs) != 2 {
		t.Fatalf("expected two inputs, got %d", len(compiled.Inputs))
	}
}

func TestCompileDefaultWorkflowUsesFallbackPath(t *testing.T) {
	t.Parallel()

	compiled, err := compileDefaultWorkflow(Config{})
	if err != nil {
		t.Fatalf("compile default workflow by fallback path: %v", err)
	}
	if compiled.Name != "dagruntime.default" {
		t.Fatalf("expected workflow name dagruntime.default, got %q", compiled.Name)
	}
}

func TestCompileSMACrossWorkflowFromYAML(t *testing.T) {
	t.Parallel()

	compiled, err := compileWorkflowFromSpecPath("workflows/sma_cross_signal.yaml")
	if err != nil {
		t.Fatalf("compile sma_cross_signal workflow: %v", err)
	}
	if compiled.Name != "dagruntime.sma_cross_signal" {
		t.Fatalf("expected workflow name dagruntime.sma_cross_signal, got %q", compiled.Name)
	}
	if len(compiled.Nodes) != 4 {
		t.Fatalf("expected four nodes, got %d", len(compiled.Nodes))
	}
	if len(compiled.Inputs) != 2 {
		t.Fatalf("expected two inputs, got %d", len(compiled.Inputs))
	}
}
