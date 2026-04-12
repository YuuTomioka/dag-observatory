package handler

import (
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	domainstate "dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type Handler struct {
	stateStore       domainstate.Store
	listTradeResults *usecase.ListTradeResults
	getSummary       *usecase.GetStrategySummary
	runWF            *usecase.RunWorkflow
	runBacktest      *usecase.RunBacktest
}

type Dependencies struct {
	StateStore         domainstate.Store
	ListTradeResults   *usecase.ListTradeResults
	GetStrategySummary *usecase.GetStrategySummary
	RunWorkflow        *usecase.RunWorkflow
	RunBacktest        *usecase.RunBacktest
}

func New(d Dependencies) *Handler {
	return &Handler{
		stateStore:       d.StateStore,
		listTradeResults: d.ListTradeResults,
		getSummary:       d.GetStrategySummary,
		runWF:            d.RunWorkflow,
		runBacktest:      d.RunBacktest,
	}
}
