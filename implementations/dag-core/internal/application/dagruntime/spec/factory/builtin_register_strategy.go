package factory

func registerStrategyFactories(registry *Registry, deps Dependencies) error {
	_ = deps
	return registerFactories(registry,
		&HeavyCalcFactory{},
		&SMAFactory{},
		&CrossDetectorFactory{},
		&SignalMapperFactory{},
		&MarketTickInputFactory{},
		&MarketBarInputM1Factory{},
		&MarketBarInputH1Factory{},
		&FeatureATRFactory{},
		&FeatureRangeHighFactory{},
		&FeatureRangeLowFactory{},
		&FeatureSpreadFactory{},
		&FeatureSessionStateFactory{},
		&FeatureHigherTFTrendFactory{},
		&SignalBreakoutLongFactory{},
		&SignalBreakoutShortFactory{},
		&SignalExitBasicFactory{},
		&FilterSessionFactory{},
		&FilterSpreadFactory{},
		&FilterEconomicEventFactory{},
		&FilterHigherTFAlignmentFactory{},
		&FilterDailyLossLimitFactory{},
		&RiskPositionSizingFactory{},
		&RiskMaxPositionsCheckFactory{},
		&RiskStopLossFromATRFactory{},
		&RiskTakeProfitFromRRFactory{},
		&SignalDecisionMapperFactory{},
	)
}
