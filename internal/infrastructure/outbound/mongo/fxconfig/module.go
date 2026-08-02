package mongo

import (
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/category"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/product"
	"github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/outbound/mongo"
	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/mongo"
	"github.com/Sokol111/ecommerce-commons/pkg/tenant"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/fx"
)

// NewMongoModule provides MongoDB infrastructure dependencies
func NewMongoModule() fx.Option {
	return fx.Provide(
		mongo.NewProductMapper,
		mongo.NewProductRepository,
		mongo.NewCategoryMapper,
		mongo.NewCategoryRepository,
		mongo.NewAttributeMapper,
		mongo.NewAttributeRepository,
		func(database *mongodriver.Database, mapper *mongo.ProductMapper) (*commonsmongo.GenericRepository[product.Product, mongo.ProductEntity], error) {
			return commonsmongo.NewGenericRepository(
				tenant.NewMultiTenantCollectionProvider(database, "product"),
				mapper,
			)
		},
		func(database *mongodriver.Database, mapper *mongo.CategoryMapper) (*commonsmongo.GenericRepository[category.Category, mongo.CategoryEntity], error) {
			return commonsmongo.NewGenericRepository(
				tenant.NewMultiTenantCollectionProvider(database, "category"),
				mapper,
			)
		},
		func(database *mongodriver.Database, mapper *mongo.AttributeMapper) (*commonsmongo.GenericRepository[attribute.Attribute, mongo.AttributeEntity], error) {
			return commonsmongo.NewGenericRepository(
				tenant.NewMultiTenantCollectionProvider(database, "attribute"),
				mapper,
			)
		},
	)
}
