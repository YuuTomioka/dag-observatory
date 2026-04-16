package di

import (
	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	marketdatarepo "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/repository/marketdata"
)

type MarketDataContainer struct {
	Client        *tsdb.Client
	Symbols       apprepository.SymbolRepository
	Ticks         apprepository.TickRepository
	TimeframeBars apprepository.TimeframeBarRepository
	SessionBars   apprepository.SessionBarRepository
	UnitOfWork    apprepository.UnitOfWork

	CreateSymbol              *marketdatausecase.CreateSymbol
	GetSymbolByCode           *marketdatausecase.GetSymbolByCode
	ListSymbols               *marketdatausecase.ListSymbols
	UpsertTicks               *marketdatausecase.UpsertTicks
	GetLatestTickBySymbol     *marketdatausecase.GetLatestTickBySymbol
	ListTicksBySymbolAndRange *marketdatausecase.ListTicksBySymbolAndRange
	BackfillTimeframeBars     *marketdatausecase.BackfillTimeframeBars
	GetLatestTimeframeBar     *marketdatausecase.GetLatestTimeframeBarBySymbolAndTimeframe
	ListTimeframeBars         *marketdatausecase.ListTimeframeBarsBySymbolTimeframeAndRange
}

func NewMarketDataContainer(cfg Config) (*MarketDataContainer, error) {
	if cfg.TSDBURL == "" {
		return nil, nil
	}

	client, err := tsdb.New(cfg.TSDBURL)
	if err != nil {
		return nil, err
	}

	uow := marketdatarepo.NewUnitOfWork(client)

	return &MarketDataContainer{
		Client:        client,
		Symbols:       marketdatarepo.NewSymbolRepository(client),
		Ticks:         marketdatarepo.NewTickRepository(client),
		TimeframeBars: marketdatarepo.NewTimeframeBarRepository(client),
		SessionBars:   marketdatarepo.NewSessionBarRepository(client),
		UnitOfWork:    uow,
		CreateSymbol: &marketdatausecase.CreateSymbol{
			UnitOfWork: uow,
		},
		GetSymbolByCode: &marketdatausecase.GetSymbolByCode{
			UnitOfWork: uow,
		},
		ListSymbols: &marketdatausecase.ListSymbols{
			UnitOfWork: uow,
		},
		UpsertTicks: &marketdatausecase.UpsertTicks{
			UnitOfWork: uow,
		},
		GetLatestTickBySymbol: &marketdatausecase.GetLatestTickBySymbol{
			UnitOfWork: uow,
		},
		ListTicksBySymbolAndRange: &marketdatausecase.ListTicksBySymbolAndRange{
			UnitOfWork: uow,
		},
		BackfillTimeframeBars: &marketdatausecase.BackfillTimeframeBars{
			UnitOfWork: uow,
		},
		GetLatestTimeframeBar: &marketdatausecase.GetLatestTimeframeBarBySymbolAndTimeframe{
			UnitOfWork: uow,
		},
		ListTimeframeBars: &marketdatausecase.ListTimeframeBarsBySymbolTimeframeAndRange{
			UnitOfWork: uow,
		},
	}, nil
}
