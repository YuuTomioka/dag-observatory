package di

import (
	graphqlif "dag-observatory/dag-core/internal/interface/graphql"
	grpcif "dag-observatory/dag-core/internal/interface/grpc"
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
	MarketData *MarketDataContainer
}

func NewAppContainer() (*AppContainer, error) {
	cfg := NewConfig()

	otelc, err := NewOTelContainer(cfg)
	if err != nil {
		return nil, err
	}

	marketData, err := NewMarketDataContainer(cfg)
	if err != nil {
		return nil, err
	}

	compiled, err := compileDefaultWorkflow(cfg, marketData)
	if err != nil {
		return nil, err
	}
	dagRuntime, err := NewDAGRuntimeContainer(cfg, otelc, compiled)
	if err != nil {
		return nil, err
	}

	e, err := NewEchoContainer(cfg, otelc, dagRuntime, marketData)
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
		MarketData: marketData,
	}, nil
}
