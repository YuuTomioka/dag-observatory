package di

import grpcif "dag-observatory/dag-core/internal/interface/grpc"

func NewGRPCContainer(cfg Config, otelc *OTelContainer, dagRuntime *DAGRuntimeContainer) (*grpcif.Server, error) {
	_ = cfg
	_ = otelc
	if dagRuntime == nil {
		return grpcif.New(grpcif.Dependencies{}), nil
	}
	return grpcif.New(grpcif.Dependencies{
		RunWorkflow: dagRuntime.Usecase,
		EventStream: dagRuntime.EventStream,
	}), nil
}
