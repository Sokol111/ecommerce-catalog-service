package product

import (
	"context"

	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

type ListQuery struct {
	Page       int
	Size       int
	Enabled    *bool
	CategoryID *string
	Sort       string
	Order      string
}

type Repository interface {
	Insert(ctx context.Context, product *Product) error

	FindByID(ctx context.Context, id string) (*Product, error)

	FindList(ctx context.Context, query ListQuery) (*commonsmongo.PageResult[Product], error)

	Update(ctx context.Context, product *Product) (*Product, error)
}
