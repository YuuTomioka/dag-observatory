package di

import (
	speccompiler "dag-observatory/dag-core/internal/application/dagruntime/spec/compiler"
	"dag-observatory/dag-core/internal/application/dagruntime/spec/factory"
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"
)

const defaultWorkflowSpecPath = "workflows/default.yaml"

func compileDefaultWorkflow(cfg Config, marketData *MarketDataContainer) (pipeline.Compiled, error) {
	specPath := cfg.WorkflowSpecPath
	if specPath == "" {
		specPath = defaultWorkflowSpecPath
	}
	return compileWorkflowFromSpecPath(specPath, marketData)
}

func compileWorkflowFromSpecPath(specPath string, marketData *MarketDataContainer) (pipeline.Compiled, error) {
	var deps factory.Dependencies
	if marketData != nil {
		deps.MarketDataUnitOfWork = marketData.UnitOfWork
	}
	registry, err := factory.NewBuiltinRegistryWithDependencies(deps)
	if err != nil {
		return pipeline.Compiled{}, err
	}
	return speccompiler.NewDefault().CompileFromPath(specPath, registry)
}
