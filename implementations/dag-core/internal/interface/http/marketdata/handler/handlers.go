package handler

import (
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
}

type Dependencies struct {
	CreateSymbol              *marketdatausecase.CreateSymbol
	GetSymbolByCode           *marketdatausecase.GetSymbolByCode
	ListSymbols               *marketdatausecase.ListSymbols
	UpsertTicks               *marketdatausecase.UpsertTicks
	GetLatestTickBySymbol     *marketdatausecase.GetLatestTickBySymbol
	ListTicksBySymbolAndRange *marketdatausecase.ListTicksBySymbolAndRange
	BackfillTimeframeBars     *marketdatausecase.BackfillTimeframeBars
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
	}
}
