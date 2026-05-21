package category

import (
	"context"
	"fmt"
)

type GetListCategoriesQuery struct {
	Page    int
	Size    int
	Enabled *bool
	Sort    string
	Order   string
}

type ListCategoriesResult struct {
	Items []*Category
	Page  int
	Size  int
	Total int64
}

type GetListCategoriesQueryHandler interface {
	Handle(ctx context.Context, query GetListCategoriesQuery) (*ListCategoriesResult, error)
}

type getListCategoriesHandler struct {
	repo Repository
}

func NewGetListCategoriesHandler(repo Repository) GetListCategoriesQueryHandler {
	return &getListCategoriesHandler{repo: repo}
}

func (h *getListCategoriesHandler) Handle(ctx context.Context, query GetListCategoriesQuery) (*ListCategoriesResult, error) {
	listQuery := ListQuery{
		Page:    query.Page,
		Size:    query.Size,
		Enabled: query.Enabled,
		Sort:    query.Sort,
		Order:   query.Order,
	}

	result, err := h.repo.FindList(ctx, listQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories list: %w", err)
	}

	return &ListCategoriesResult{
		Items: result.Items,
		Page:  result.Page,
		Size:  result.Size,
		Total: result.Total,
	}, nil
}
