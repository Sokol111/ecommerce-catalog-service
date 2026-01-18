package product

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
)

type GetProductByIDQuery struct {
	ID string
}

type GetProductByIDQueryHandler interface {
	Handle(ctx context.Context, query GetProductByIDQuery) (*product.Product, error)
}

type getProductByIDHandler struct {
	repo product.Repository
}

func NewGetProductByIDHandler(repo product.Repository) GetProductByIDQueryHandler {
	return &getProductByIDHandler{repo: repo}
}

func (h *getProductByIDHandler) Handle(ctx context.Context, query GetProductByIDQuery) (*product.Product, error) {
	p, err := h.repo.FindByID(ctx, query.ID)
	if err != nil {
		if errors.Is(err, persistence.ErrEntityNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return p, nil
}
