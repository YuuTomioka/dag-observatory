package handler

import (
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	domainstate "dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type Handler struct {
	stateStore       domainstate.Store
	listTradeResults *usecase.ListTradeResults
	runWF            *usecase.RunWorkflow
}

type Dependencies struct {
	StateStore       domainstate.Store
	ListTradeResults *usecase.ListTradeResults
	RunWorkflow      *usecase.RunWorkflow
}

func New(d Dependencies) *Handler {
	return &Handler{
		stateStore:       d.StateStore,
		listTradeResults: d.ListTradeResults,
		runWF:            d.RunWorkflow,
	}
}
