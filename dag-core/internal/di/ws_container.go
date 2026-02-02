package di

import wsif "dag-observatory/dag-core/internal/interface/ws"

func NewWSContainer(cfg Config, otelc *OTelContainer, dagRuntime *DAGRuntimeContainer) (*wsif.Server, error) {
	_ = cfg
	_ = otelc
	if dagRuntime == nil {
		return wsif.New(wsif.Dependencies{}), nil
	}
	return wsif.New(wsif.Dependencies{
		RunWorkflow: dagRuntime.Usecase,
		EventStream: dagRuntime.EventStream,
	}), nil
}
