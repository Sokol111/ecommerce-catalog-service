package mongo

import (
	"context"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type attributeRepository struct {
	*commonsmongo.GenericRepository[attribute.Attribute, attributeEntity]
	collection commonsmongo.Collection
}

func newAttributeRepository(mongoClient commonsmongo.Mongo, mapper *attributeMapper) (attribute.Repository, error) {
	collection := mongoClient.GetCollection("attribute")

	genericRepo, err := commonsmongo.NewGenericRepository(
		collection,
		mapper,
	)
	if err != nil {
		return nil, err
	}

	return &attributeRepository{
		GenericRepository: genericRepo,
		collection:        collection,
	}, nil
}

func (r *attributeRepository) FindList(ctx context.Context, query attribute.ListQuery) (*commonsmongo.PageResult[attribute.Attribute], error) {
	filter := bson.D{}
	if query.Enabled != nil {
		filter = append(filter, bson.E{Key: "enabled", Value: *query.Enabled})
	}
	if query.Type != nil {
		filter = append(filter, bson.E{Key: "type", Value: *query.Type})
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

// Override Insert to handle duplicate slug error
func (r *attributeRepository) Insert(ctx context.Context, a *attribute.Attribute) error {
	err := r.GenericRepository.Insert(ctx, a)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return attribute.ErrSlugAlreadyExists
		}
		return err
	}
	return nil
}

// Override Update to handle duplicate slug error
func (r *attributeRepository) Update(ctx context.Context, a *attribute.Attribute) (*attribute.Attribute, error) {
	result, err := r.GenericRepository.Update(ctx, a)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, attribute.ErrSlugAlreadyExists
		}
		return nil, err
	}
	return result, nil
}
