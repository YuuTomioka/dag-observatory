package validator

import (
	"fmt"
	"testing"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
)

type stubKinds map[string]bool

func (k stubKinds) Has(kind string) bool {
	return k[kind]
}

type validatingKinds struct {
	known   map[string]bool
	invalid map[string]error
}

func (k validatingKinds) Has(kind string) bool {
	return k.known[kind]
}

func (k validatingKinds) ValidateNodeSpec(nodeSpec spec.NodeSpec) error {
	if err, ok := k.invalid[nodeSpec.ID]; ok {
		return err
	}
	return nil
}

func TestValidateWorkflowSpecSuccess(t *testing.T) {
	t.Parallel()

	err := ValidateWorkflowSpec(
		spec.WorkflowSpec{
			Name:    "dagruntime.default",
			Version: "v1",
			Nodes: []spec.NodeSpec{
				{ID: "heavy_calc", Kind: "heavy_calc", Config: map[string]any{}},
			},
		},
		stubKinds{"heavy_calc": true},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateWorkflowSpecUnknownKind(t *testing.T) {
	t.Parallel()

	err := ValidateWorkflowSpec(
		spec.WorkflowSpec{
			Name:    "dagruntime.default",
			Version: "v1",
			Nodes: []spec.NodeSpec{
				{ID: "heavy_calc", Kind: "unknown_kind"},
			},
		},
		stubKinds{"heavy_calc": true},
	)
	if err == nil {
		t.Fatal("expected unknown kind error")
	}
}

func TestValidateWorkflowSpecDuplicateNodeID(t *testing.T) {
	t.Parallel()

	err := ValidateWorkflowSpec(
		spec.WorkflowSpec{
			Name:    "dagruntime.default",
			Version: "v1",
			Nodes: []spec.NodeSpec{
				{ID: "dup", Kind: "heavy_calc"},
				{ID: "dup", Kind: "heavy_calc"},
			},
		},
		stubKinds{"heavy_calc": true},
	)
	if err == nil {
		t.Fatal("expected duplicate node id error")
	}
}

func TestValidateWorkflowSpecUnsupportedVersion(t *testing.T) {
	t.Parallel()

	err := ValidateWorkflowSpec(
		spec.WorkflowSpec{
			Name:    "dagruntime.default",
			Version: "v2",
			Nodes: []spec.NodeSpec{
				{ID: "n1", Kind: "heavy_calc"},
			},
		},
		stubKinds{"heavy_calc": true},
	)
	if err == nil {
		t.Fatal("expected unsupported version error")
	}
}

func TestValidateWorkflowSpecConfigValidation(t *testing.T) {
	t.Parallel()

	err := ValidateWorkflowSpec(
		spec.WorkflowSpec{
			Name:    "dagruntime.default",
			Version: "v1",
			Nodes: []spec.NodeSpec{
				{ID: "n1", Kind: "sma", Config: map[string]any{"window": -1}},
			},
		},
		validatingKinds{
			known:   map[string]bool{"sma": true},
			invalid: map[string]error{"n1": fmt.Errorf("window must be > 0")},
		},
	)
	if err == nil {
		t.Fatal("expected config validation error")
	}
}
