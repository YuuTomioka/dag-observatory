package validator

import (
	"fmt"
	"strings"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
)

type KindLookup interface {
	Has(kind string) bool
}

type NodeConfigValidator interface {
	ValidateNodeSpec(nodeSpec spec.NodeSpec) error
}

func ValidateWorkflowSpec(wfSpec spec.WorkflowSpec, kinds KindLookup) error {
	if strings.TrimSpace(wfSpec.Name) == "" {
		return fmt.Errorf("dagruntime spec validator: name is required")
	}
	if err := validateVersion(wfSpec.Version); err != nil {
		return err
	}
	if len(wfSpec.Nodes) == 0 {
		return fmt.Errorf("dagruntime spec validator: at least one node is required")
	}

	seenNodeIDs := make(map[string]struct{}, len(wfSpec.Nodes))
	for i, nodeSpec := range wfSpec.Nodes {
		nodeID := strings.TrimSpace(nodeSpec.ID)
		if nodeID == "" {
			return fmt.Errorf("dagruntime spec validator: nodes[%d].id is required", i)
		}
		if _, ok := seenNodeIDs[nodeID]; ok {
			return fmt.Errorf("dagruntime spec validator: duplicate node id %q", nodeID)
		}
		seenNodeIDs[nodeID] = struct{}{}

		kind := strings.TrimSpace(nodeSpec.Kind)
		if kind == "" {
			return fmt.Errorf("dagruntime spec validator: nodes[%d].kind is required", i)
		}
		if kinds != nil && !kinds.Has(kind) {
			return fmt.Errorf("dagruntime spec validator: nodes[%d].kind %q is not registered", i, kind)
		}
		if configValidator, ok := kinds.(NodeConfigValidator); ok {
			if err := configValidator.ValidateNodeSpec(nodeSpec); err != nil {
				return fmt.Errorf("dagruntime spec validator: nodes[%d] config validation failed: %w", i, err)
			}
		}
		if err := validateNodeConfig(nodeSpec.Config, i); err != nil {
			return err
		}
	}

	return nil
}

func validateVersion(version string) error {
	version = strings.TrimSpace(version)
	if version == "" {
		return fmt.Errorf("dagruntime spec validator: version is required")
	}
	if strings.ContainsAny(version, " \t\r\n") {
		return fmt.Errorf("dagruntime spec validator: version %q is invalid", version)
	}
	if version != "v1" {
		return fmt.Errorf("dagruntime spec validator: version %q is not supported (supported: v1)", version)
	}
	return nil
}

func validateNodeConfig(config map[string]any, nodeIndex int) error {
	if config == nil {
		return nil
	}
	for key := range config {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("dagruntime spec validator: nodes[%d].config contains empty key", nodeIndex)
		}
	}
	return nil
}
