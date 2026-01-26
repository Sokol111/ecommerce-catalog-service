package category

import (
	"context"

	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

type ListQuery struct {
	Page    int
	Size    int
	Enabled *bool
	Sort    string
	Order   string
}

type Repository interface {
	Insert(ctx context.Context, category *Category) error

	FindByID(ctx context.Context, id string) (*Category, error)

	FindList(ctx context.Context, query ListQuery) (*commonsmongo.PageResult[Category], error)

	Update(ctx context.Context, category *Category) (*Category, error)

	Exists(ctx context.Context, id string) (bool, error)
}
