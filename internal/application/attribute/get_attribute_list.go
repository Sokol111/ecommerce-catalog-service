package attribute

import (
	"context"
	"fmt"
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
	Items []*Attribute
	Page  int
	Size  int
	Total int64
}

type GetAttributeListQueryHandler interface {
	Handle(ctx context.Context, query GetAttributeListQuery) (*ListAttributesResult, error)
}

type getAttributeListHandler struct {
	repo Repository
}

func NewGetAttributeListHandler(repo Repository) GetAttributeListQueryHandler {
	return &getAttributeListHandler{repo: repo}
}

func (h *getAttributeListHandler) Handle(ctx context.Context, query GetAttributeListQuery) (*ListAttributesResult, error) {
	listQuery := ListQuery(query)

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
