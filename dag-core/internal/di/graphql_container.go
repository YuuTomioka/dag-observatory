package di

import graphqlif "dag-observatory/dag-core/internal/interface/graphql"

func NewGraphQLContainer(cfg Config, otelc *OTelContainer, dagRuntime *DAGRuntimeContainer) (*graphqlif.Server, error) {
	_ = cfg
	_ = otelc
	if dagRuntime == nil {
		return graphqlif.New(graphqlif.Dependencies{}), nil
	}
	return graphqlif.New(graphqlif.Dependencies{RunWorkflow: dagRuntime.Usecase}), nil
}
