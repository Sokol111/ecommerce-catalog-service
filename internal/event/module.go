package event

import "go.uber.org/fx"

// EventModule provides event factory dependencies
func EventModule() fx.Option {
	return fx.Options(
		fx.Provide(
			newProductEventFactory,
			newCategoryEventFactory,
			newAttributeEventFactory,
		),
	)
}
