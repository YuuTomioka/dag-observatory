package factory

func registerMarketdataFactories(registry *Registry, deps Dependencies) error {
	if deps.MarketDataUnitOfWork == nil {
		return nil
	}
	return registerFactories(registry,
		&TimeframeBarBackfillFactory{UnitOfWork: deps.MarketDataUnitOfWork},
		&ResolveSymbolFactory{UnitOfWork: deps.MarketDataUnitOfWork},
		&PlanWindowsFactory{},
		&LoadTicksForChunkFactory{UnitOfWork: deps.MarketDataUnitOfWork},
		&AggregateTimeframeBarsFactory{},
		&PersistTimeframeBarsFactory{UnitOfWork: deps.MarketDataUnitOfWork},
	)
}
