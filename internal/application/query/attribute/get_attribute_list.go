package query

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
)

type GetAttributeListQuery struct {
	Page    int
	Size    int
	Enabled *bool
	Type    *string
	Sort    string
	Order   string
}

type ListAttributesResult struct {
	Items []*attribute.Attribute
	Page  int
	Size  int
	Total int64
}

type GetAttributeListQueryHandler interface {
	Handle(ctx context.Context, query GetAttributeListQuery) (*ListAttributesResult, error)
}

type getAttributeListHandler struct {
	repo attribute.Repository
}

func NewGetAttributeListHandler(repo attribute.Repository) GetAttributeListQueryHandler {
	return &getAttributeListHandler{repo: repo}
}

func (h *getAttributeListHandler) Handle(ctx context.Context, query GetAttributeListQuery) (*ListAttributesResult, error) {
	listQuery := attribute.ListQuery{
		Page:    query.Page,
		Size:    query.Size,
		Enabled: query.Enabled,
		Type:    query.Type,
		Sort:    query.Sort,
		Order:   query.Order,
	}

	result, err := h.repo.FindList(ctx, listQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get attributes list: %w", err)
	}

	return &ListAttributesResult{
		Items: result.Items,
		Page:  result.Page,
		Size:  result.Size,
		Total: result.Total,
	}, nil
}
