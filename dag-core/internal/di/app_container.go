package di

import (
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"
	grpcif "dag-observatory/dag-core/internal/interface/grpc"
	graphqlif "dag-observatory/dag-core/internal/interface/graphql"
	wsif "dag-observatory/dag-core/internal/interface/ws"

	"github.com/labstack/echo/v4"
)

type AppContainer struct {
	Config     Config
	OTel       *OTelContainer
	Echo       *echo.Echo
	GRPC       *grpcif.Server
	GraphQL    *graphqlif.Server
	WS         *wsif.Server
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

	grpcServer, err := NewGRPCContainer(cfg, otelc, dagRuntime)
	if err != nil {
		return nil, err
	}

	graphqlServer, err := NewGraphQLContainer(cfg, otelc, dagRuntime)
	if err != nil {
		return nil, err
	}

	wsServer, err := NewWSContainer(cfg, otelc, dagRuntime)
	if err != nil {
		return nil, err
	}

	return &AppContainer{
		Config:     cfg,
		OTel:       otelc,
		Echo:       e,
		GRPC:       grpcServer,
		GraphQL:    graphqlServer,
		WS:         wsServer,
		DAGRuntime: dagRuntime,
	}, nil
}
