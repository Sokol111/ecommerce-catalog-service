package attribute

import (
	"context"

	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

type ListQuery struct {
	Page    int
	Size    int
	Enabled *bool
	Type    *string
	Sort    string
	Order   string
}

type Repository interface {
	Insert(ctx context.Context, attribute *Attribute) error

	FindByID(ctx context.Context, id string) (*Attribute, error)

	FindByIDs(ctx context.Context, ids []string) ([]*Attribute, error)

	// FindByIDsOrFail returns attributes by IDs or error if any ID is not found
	FindByIDsOrFail(ctx context.Context, ids []string) ([]*Attribute, error)

	FindList(ctx context.Context, query ListQuery) (*commonsmongo.PageResult[Attribute], error)

	Update(ctx context.Context, attribute *Attribute) (*Attribute, error)

	Exists(ctx context.Context, id string) (bool, error)
}
