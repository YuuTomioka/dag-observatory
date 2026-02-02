package graphql

import "dag-observatory/dag-core/internal/application/dagruntime/usecase"

// Server is a placeholder for future GraphQL transport.
type Server struct {
	runWorkflow *usecase.RunWorkflow
}

type Dependencies struct {
	RunWorkflow *usecase.RunWorkflow
}

func New(d Dependencies) *Server {
	return &Server{
		runWorkflow: d.RunWorkflow,
	}
}
