package category

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
)

type GetCategoryByIDQuery struct {
	ID string
}

type GetCategoryByIDQueryHandler interface {
	Handle(ctx context.Context, query GetCategoryByIDQuery) (*category.Category, error)
}

type getCategoryByIDHandler struct {
	repo category.Repository
}

func NewGetCategoryByIDHandler(repo category.Repository) GetCategoryByIDQueryHandler {
	return &getCategoryByIDHandler{repo: repo}
}

func (h *getCategoryByIDHandler) Handle(ctx context.Context, query GetCategoryByIDQuery) (*category.Category, error) {
	c, err := h.repo.FindByID(ctx, query.ID)
	if err != nil {
		if errors.Is(err, persistence.ErrEntityNotFound) {
			return nil, persistence.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return c, nil
}
