package mongo

import (
	"context"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type productRepository struct {
	*commonsmongo.GenericRepository[product.Product, productEntity]
}

func newProductRepository(mongo commonsmongo.Mongo, mapper *productMapper) (product.Repository, error) {
	genericRepo, err := commonsmongo.NewGenericRepository(
		mongo.GetCollection("product"),
		mapper,
	)
	if err != nil {
		return nil, err
	}

	return &productRepository{
		GenericRepository: genericRepo,
	}, nil
}

func (r *productRepository) FindList(ctx context.Context, query product.ListQuery) (*commonsmongo.PageResult[product.Product], error) {
	filter := bson.D{}
	if query.Enabled != nil {
		filter = append(filter, bson.E{Key: "enabled", Value: *query.Enabled})
	}
	if query.CategoryID != nil {
		filter = append(filter, bson.E{Key: "categoryId", Value: *query.CategoryID})
	}

	var sortBson bson.D
	if query.Sort != "" {
		sortOrder := 1 // asc
		if query.Order == "desc" {
			sortOrder = -1
		}
		sortBson = bson.D{{Key: query.Sort, Value: sortOrder}}
	}

	opts := commonsmongo.QueryOptions{
		Filter: filter,
		Page:   query.Page,
		Size:   query.Size,
		Sort:   sortBson,
	}

	return r.FindWithOptions(ctx, opts)
}
