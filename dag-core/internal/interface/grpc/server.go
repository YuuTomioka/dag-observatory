package grpc

import (
	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
)

// Server is a placeholder for future gRPC transport.
type Server struct {
	runWorkflow *usecase.RunWorkflow
	eventStream port.EventStream
}

type Dependencies struct {
	RunWorkflow *usecase.RunWorkflow
	EventStream port.EventStream
}

func New(d Dependencies) *Server {
	return &Server{
		runWorkflow: d.RunWorkflow,
		eventStream: d.EventStream,
	}
}
