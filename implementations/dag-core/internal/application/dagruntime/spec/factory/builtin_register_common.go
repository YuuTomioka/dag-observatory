package factory

func registerFactories(registry *Registry, factories ...NodeFactory) error {
	for _, f := range factories {
		if err := registry.Register(f); err != nil {
			return err
		}
	}
	return nil
}
