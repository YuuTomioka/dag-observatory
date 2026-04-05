package factory

func registerPositionFactories(registry *Registry, deps Dependencies) error {
	_ = deps
	return registerFactories(registry,
		&PositionSnapshotLoadFactory{},
		&PositionBreakevenFactory{},
		&PositionTrailingStopFactory{},
		&PositionTimeoutExitFactory{},
		&PositionTrackerUpdateFactory{},
		&PositionCloseToClosedTradeFactory{},
		&ClosedTradeStoreFactory{},
		&DailyPnLUpdateFactory{},
		&OpenPositionCloseFactory{},
	)
}
