package handler

import (
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	domainstate "dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type Handler struct {
	stateStore       domainstate.Store
	listTradeResults *usecase.ListTradeResults
	getEquitySeries  *usecase.GetEquitySeries
	getRun           *usecase.GetRun
	getSummary       *usecase.GetStrategySummary
	runWF            *usecase.RunWorkflow
	runBacktest      *usecase.RunBacktest
}

type Dependencies struct {
	StateStore         domainstate.Store
	ListTradeResults   *usecase.ListTradeResults
	GetEquitySeries    *usecase.GetEquitySeries
	GetRun             *usecase.GetRun
	GetStrategySummary *usecase.GetStrategySummary
	RunWorkflow        *usecase.RunWorkflow
	RunBacktest        *usecase.RunBacktest
}

func New(d Dependencies) *Handler {
	return &Handler{
		stateStore:       d.StateStore,
		listTradeResults: d.ListTradeResults,
		getEquitySeries:  d.GetEquitySeries,
		getRun:           d.GetRun,
		getSummary:       d.GetStrategySummary,
		runWF:            d.RunWorkflow,
		runBacktest:      d.RunBacktest,
	}
}
