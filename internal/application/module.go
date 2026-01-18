package application

import (
	attrcommand "github.com/Sokol111/ecommerce-catalog-service/internal/application/command/attribute"
	catcommand "github.com/Sokol111/ecommerce-catalog-service/internal/application/command/category"
	prodcommand "github.com/Sokol111/ecommerce-catalog-service/internal/application/command/product"
	attrquery "github.com/Sokol111/ecommerce-catalog-service/internal/application/query/attribute"
	catquery "github.com/Sokol111/ecommerce-catalog-service/internal/application/query/category"
	prodquery "github.com/Sokol111/ecommerce-catalog-service/internal/application/query/product"
	"go.uber.org/fx"
)

// Module provides application layer dependencies
func Module() fx.Option {
	return fx.Options(
		// Command handlers
		fx.Provide(
			prodcommand.NewCreateProductHandler,
			prodcommand.NewUpdateProductHandler,
			catcommand.NewCreateCategoryHandler,
			catcommand.NewUpdateCategoryHandler,
			attrcommand.NewCreateAttributeHandler,
			attrcommand.NewUpdateAttributeHandler,
		),
		// Query handlers
		fx.Provide(
			prodquery.NewGetProductByIDHandler,
			prodquery.NewGetListProductsHandler,
			catquery.NewGetCategoryByIDHandler,
			catquery.NewGetListCategoriesHandler,
			attrquery.NewGetAttributeByIDHandler,
			attrquery.NewGetAttributeListHandler,
		),
	)
}
