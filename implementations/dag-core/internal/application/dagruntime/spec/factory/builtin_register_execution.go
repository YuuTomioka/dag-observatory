package factory

func registerExecutionFactories(registry *Registry, deps Dependencies) error {
	_ = deps
	return registerFactories(registry,
		&ExecutionSubmitPaperOrderFactory{},
		&ExecutionSubmitMarketOrderFactory{},
		&ExecutionConfirmFillFactory{},
	)
}
