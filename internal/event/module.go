package event

import "go.uber.org/fx"

// Module provides event factory dependencies
func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			newProductEventFactory,
			newCategoryEventFactory,
			newAttributeEventFactory,
		),
	)
}
