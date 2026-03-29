package di

import (
	"fmt"
	"os"
	"path/filepath"

	"dag-observatory/dag-core/internal/application/dagruntime/spec/adapter"
	"dag-observatory/dag-core/internal/application/dagruntime/spec/factory"
	"dag-observatory/dag-core/internal/application/dagruntime/spec/loader"
	"dag-observatory/dag-core/internal/application/dagruntime/spec/validator"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"
	"dag-observatory/dag-core/internal/domain/dagruntime/workflow"
)

const defaultWorkflowSpecPath = "workflows/default.yaml"

func compileDefaultWorkflow(cfg Config) (pipeline.Compiled, error) {
	specPath := cfg.WorkflowSpecPath
	if specPath == "" {
		specPath = defaultWorkflowSpecPath
	}
	return compileWorkflowFromSpecPath(specPath)
}

func compileWorkflowFromSpecPath(specPath string) (pipeline.Compiled, error) {
	registry, err := factory.NewBuiltinRegistry()
	if err != nil {
		return pipeline.Compiled{}, err
	}

	workflowLoader := loader.Loader{}
	wfSpecPath, err := resolveWorkflowSpecPath(specPath)
	if err != nil {
		return pipeline.Compiled{}, err
	}
	wfSpec, err := workflowLoader.LoadFile(wfSpecPath)
	if err != nil {
		return pipeline.Compiled{}, err
	}
	if err := validator.ValidateWorkflowSpec(wfSpec, registry); err != nil {
		return pipeline.Compiled{}, err
	}

	wf, err := adapter.ToWorkflow(
		wfSpec,
		map[string]artifact.AnyKey{
			"symbol":        usecase.InputKeySymbol,
			"mode":          usecase.InputKeyMode,
			"market.symbol": usecase.InputKeySymbol,
			"market.bars":   usecase.InputKeyMarketBars,
		},
		registry,
	)
	if err != nil {
		return pipeline.Compiled{}, err
	}
	return workflow.Compile(wf)
}

func resolveWorkflowSpecPath(path string) (string, error) {
	candidates := []string{
		path,
		filepath.Join("..", "..", path),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("dagruntime workflow spec not found: %s", path)
}
