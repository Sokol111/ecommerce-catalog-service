package application

import (
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/category"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/product"
	"go.uber.org/fx"
)

// NewAppModule provides application layer dependencies
func NewAppModule() fx.Option {
	return fx.Options(
		// Command handlers
		fx.Provide(
			product.NewCreateProductHandler,
			product.NewUpdateProductHandler,
			product.NewDeleteProductHandler,
			category.NewCreateCategoryHandler,
			category.NewUpdateCategoryHandler,
			attribute.NewCreateAttributeHandler,
			attribute.NewUpdateAttributeHandler,
		),
		// Query handlers
		fx.Provide(
			product.NewGetProductByIDHandler,
			product.NewGetListProductsHandler,
			category.NewGetCategoryByIDHandler,
			category.NewGetListCategoriesHandler,
			attribute.NewGetAttributeByIDHandler,
			attribute.NewGetAttributeListHandler,
		),
	)
}
