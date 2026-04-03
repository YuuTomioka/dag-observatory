package factory

func registerObservabilityFactories(registry *Registry, deps Dependencies) error {
	_ = deps
	return registerFactories(registry,
		&ObservabilityEmitSignalDecisionFactory{},
		&ObservabilityEmitOrderDecisionFactory{},
		&ObservabilityEmitPositionEventFactory{},
	)
}
