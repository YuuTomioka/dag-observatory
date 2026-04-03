package compiler

import (
	"fmt"

	"dag-observatory/dag-core/internal/application/dagruntime/spec/adapter"
	"dag-observatory/dag-core/internal/application/dagruntime/spec/factory"
	"dag-observatory/dag-core/internal/application/dagruntime/spec/inputmap"
	"dag-observatory/dag-core/internal/application/dagruntime/spec/loader"
	"dag-observatory/dag-core/internal/application/dagruntime/spec/validator"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"
	"dag-observatory/dag-core/internal/domain/dagruntime/workflow"
)

type Compiler struct {
	Loader          loader.Loader
	ResolveSpecPath func(string) (string, error)
	InputMap        map[string]artifact.AnyKey
}

func NewDefault() Compiler {
	return Compiler{
		Loader:          loader.Loader{},
		ResolveSpecPath: ResolveSpecPath,
		InputMap:        inputmap.Default(),
	}
}

func (c Compiler) CompileFromPath(specPath string, registry *factory.Registry) (pipeline.Compiled, error) {
	if registry == nil {
		return pipeline.Compiled{}, fmt.Errorf("dagruntime compiler: registry is required")
	}
	resolve := c.ResolveSpecPath
	if resolve == nil {
		resolve = func(path string) (string, error) { return path, nil }
	}
	inputs := c.InputMap
	if inputs == nil {
		inputs = inputmap.Default()
	}

	wfSpecPath, err := resolve(specPath)
	if err != nil {
		return pipeline.Compiled{}, err
	}
	wfSpec, err := c.Loader.LoadFile(wfSpecPath)
	if err != nil {
		return pipeline.Compiled{}, err
	}
	if err := validator.ValidateWorkflowSpec(wfSpec, registry); err != nil {
		return pipeline.Compiled{}, err
	}
	wf, err := adapter.ToWorkflow(wfSpec, inputs, registry)
	if err != nil {
		return pipeline.Compiled{}, err
	}
	return workflow.Compile(wf)
}
