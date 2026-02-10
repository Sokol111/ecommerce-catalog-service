package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

type GetAttributeByIDQuery struct {
	ID string
}

type GetAttributeByIDQueryHandler interface {
	Handle(ctx context.Context, query GetAttributeByIDQuery) (*attribute.Attribute, error)
}

type getAttributeByIDHandler struct {
	repo attribute.Repository
}

func NewGetAttributeByIDHandler(repo attribute.Repository) GetAttributeByIDQueryHandler {
	return &getAttributeByIDHandler{repo: repo}
}

func (h *getAttributeByIDHandler) Handle(ctx context.Context, query GetAttributeByIDQuery) (*attribute.Attribute, error) {
	a, err := h.repo.FindByID(ctx, query.ID)
	if err != nil {
		if errors.Is(err, mongo.ErrEntityNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}
	return a, nil
}
