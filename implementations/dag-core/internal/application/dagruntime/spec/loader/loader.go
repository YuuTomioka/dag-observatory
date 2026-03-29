package loader

import (
	"fmt"
	"os"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"

	"gopkg.in/yaml.v3"
)

type Loader struct{}

func (l Loader) LoadFile(path string) (spec.WorkflowSpec, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return spec.WorkflowSpec{}, fmt.Errorf("dagruntime spec loader: read %s: %w", path, err)
	}

	var wfSpec spec.WorkflowSpec
	if err := yaml.Unmarshal(raw, &wfSpec); err != nil {
		return spec.WorkflowSpec{}, fmt.Errorf("dagruntime spec loader: decode %s: %w", path, err)
	}
	return wfSpec, nil
}
