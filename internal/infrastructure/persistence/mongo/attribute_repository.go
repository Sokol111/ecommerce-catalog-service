package mongo

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type attributeRepository struct {
	*commonsmongo.GenericRepository[attribute.Attribute, attributeEntity]
}

func newAttributeRepository(mongoClient commonsmongo.Mongo, mapper *attributeMapper) (attribute.Repository, error) {
	genericRepo, err := commonsmongo.NewGenericRepository(
		mongoClient.GetCollection("attribute"),
		mapper,
	)
	if err != nil {
		return nil, err
	}

	return &attributeRepository{
		GenericRepository: genericRepo,
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

func (r *attributeRepository) FindByIDs(ctx context.Context, ids []string) ([]*attribute.Attribute, error) {
	if len(ids) == 0 {
		return []*attribute.Attribute{}, nil
	}

	filter := bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}}
	return r.FindAllWithFilter(ctx, filter, nil)
}

func (r *attributeRepository) FindByIDsOrFail(ctx context.Context, ids []string) ([]*attribute.Attribute, error) {
	attrs, err := r.FindByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attributes: %w", err)
	}

	if len(attrs) != len(ids) {
		foundIDs := lo.SliceToMap(attrs, func(a *attribute.Attribute) (string, struct{}) {
			return a.ID, struct{}{}
		})
		missingID, _ := lo.Find(ids, func(id string) bool {
			_, exists := foundIDs[id]
			return !exists
		})
		return nil, fmt.Errorf("attribute not found: %s", missingID)
	}

	return attrs, nil
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
