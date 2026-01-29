package di

import (
	"dag-observatory/demo-go/internal/domain/dagruntime/driver"
	"dag-observatory/demo-go/internal/domain/dagruntime/engine"
	"dag-observatory/demo-go/internal/domain/dagruntime/pipeline"
	"dag-observatory/demo-go/internal/domain/dagruntime/policy"
	artifactinfra "dag-observatory/demo-go/internal/infrastructure/dagruntime/artifact"
	observerinfra "dag-observatory/demo-go/internal/infrastructure/dagruntime/observer"
	recorderinfra "dag-observatory/demo-go/internal/infrastructure/dagruntime/recorder"
	stateinfra "dag-observatory/demo-go/internal/infrastructure/dagruntime/state"
)

type DAGRuntimeContainer struct {
	ArtifactStore *artifactinfra.MemoryStore
	StateStore    *stateinfra.MemoryStore
	Observer      *observerinfra.OTelObserver
	Recorder      *recorderinfra.NoopRecorder
	Runner        *engine.Runner
	Driver        *driver.Driver
}

func NewDAGRuntimeContainer(compiled pipeline.Compiled) *DAGRuntimeContainer {
	artifactStore := artifactinfra.NewMemoryStore()
	stateStore := stateinfra.NewMemoryStore()
	observer := observerinfra.NewOTelObserver()
	recorder := recorderinfra.NewNoopRecorder()

	runner := &engine.Runner{
		ArtifactStore: artifactStore,
		StateStore:    stateStore,
		Observer:      observer,
		Recorder:      recorder,
		Policy: policy.Policy{
			DefaultRetry: policy.RetryPolicy{MaxAttempts: 1},
		},
	}

	driver := &driver.Driver{
		Runner:   runner,
		Compiled: compiled,
	}

	return &DAGRuntimeContainer{
		ArtifactStore: artifactStore,
		StateStore:    stateStore,
		Observer:      observer,
		Recorder:      recorder,
		Runner:        runner,
		Driver:        driver,
	}
}

