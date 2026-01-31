package di

import (
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"

	"github.com/labstack/echo/v4"
)

type AppContainer struct {
	Config     Config
	OTel       *OTelContainer
	Echo       *echo.Echo
	DAGRuntime *DAGRuntimeContainer
}

func NewAppContainer() (*AppContainer, error) {
	cfg := NewConfig()

	otelc, err := NewOTelContainer(cfg)
	if err != nil {
		return nil, err
	}

	compiled := pipeline.Compiled{}
	dagRuntime, err := NewDAGRuntimeContainer(cfg, otelc, compiled)
	if err != nil {
		return nil, err
	}

	e, err := NewEchoContainer(cfg, otelc, dagRuntime)
	if err != nil {
		return nil, err
	}

	return &AppContainer{
		Config:     cfg,
		OTel:       otelc,
		Echo:       e,
		DAGRuntime: dagRuntime,
	}, nil
}
