package attribute

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

type GetAttributeByIDQuery struct {
	ID string
}

type GetAttributeByIDQueryHandler interface {
	Handle(ctx context.Context, query GetAttributeByIDQuery) (*Attribute, error)
}

type getAttributeByIDHandler struct {
	repo Repository
}

func NewGetAttributeByIDHandler(repo Repository) GetAttributeByIDQueryHandler {
	return &getAttributeByIDHandler{repo: repo}
}

func (h *getAttributeByIDHandler) Handle(ctx context.Context, query GetAttributeByIDQuery) (*Attribute, error) {
	a, err := h.repo.FindByID(ctx, query.ID)
	if err != nil {
		if errors.Is(err, mongo.ErrEntityNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}
	return a, nil
}
