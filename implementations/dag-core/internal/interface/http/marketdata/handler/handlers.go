package handler

import (
	dagruntimeusecase "dag-observatory/dag-core/internal/application/dagruntime/usecase"
	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
)

type Handlers struct {
	createSymbol              *marketdatausecase.CreateSymbol
	getSymbolByCode           *marketdatausecase.GetSymbolByCode
	listSymbols               *marketdatausecase.ListSymbols
	upsertTicks               *marketdatausecase.UpsertTicks
	getLatestTickBySymbol     *marketdatausecase.GetLatestTickBySymbol
	listTicksBySymbolAndRange *marketdatausecase.ListTicksBySymbolAndRange
	backfillTimeframeBars     *marketdatausecase.BackfillTimeframeBars
	runWorkflow               *dagruntimeusecase.RunWorkflow
}

type Dependencies struct {
	CreateSymbol              *marketdatausecase.CreateSymbol
	GetSymbolByCode           *marketdatausecase.GetSymbolByCode
	ListSymbols               *marketdatausecase.ListSymbols
	UpsertTicks               *marketdatausecase.UpsertTicks
	GetLatestTickBySymbol     *marketdatausecase.GetLatestTickBySymbol
	ListTicksBySymbolAndRange *marketdatausecase.ListTicksBySymbolAndRange
	BackfillTimeframeBars     *marketdatausecase.BackfillTimeframeBars
	RunWorkflow               *dagruntimeusecase.RunWorkflow
}

func New(d Dependencies) *Handlers {
	return &Handlers{
		createSymbol:              d.CreateSymbol,
		getSymbolByCode:           d.GetSymbolByCode,
		listSymbols:               d.ListSymbols,
		upsertTicks:               d.UpsertTicks,
		getLatestTickBySymbol:     d.GetLatestTickBySymbol,
		listTicksBySymbolAndRange: d.ListTicksBySymbolAndRange,
		backfillTimeframeBars:     d.BackfillTimeframeBars,
		runWorkflow:               d.RunWorkflow,
	}
}
