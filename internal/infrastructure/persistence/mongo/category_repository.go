package mongo

import (
	"context"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type categoryRepository struct {
	*commonsmongo.GenericRepository[category.Category, categoryEntity]
}

func newCategoryRepository(mongo commonsmongo.Mongo, mapper *categoryMapper) (category.Repository, error) {
	genericRepo, err := commonsmongo.NewGenericRepository(
		mongo.GetCollection("category"),
		mapper,
	)
	if err != nil {
		return nil, err
	}

	return &categoryRepository{
		GenericRepository: genericRepo,
	}, nil
}

func (r *categoryRepository) FindList(ctx context.Context, query category.ListQuery) (*commonsmongo.PageResult[category.Category], error) {
	// Build filter
	filter := bson.D{}
	if query.Enabled != nil {
		filter = append(filter, bson.E{Key: "enabled", Value: *query.Enabled})
	}

	// Build sort
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
