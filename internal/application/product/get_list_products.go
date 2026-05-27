package product

import (
	"context"
	"fmt"
)

type GetListProductsQuery struct {
	Page       int
	Size       int
	Enabled    *bool
	CategoryID *string
	Sort       string
	Order      string
}

type ListProductsResult struct {
	Items []*Product
	Page  int
	Size  int
	Total int64
}

type GetListProductsQueryHandler interface {
	Handle(ctx context.Context, query GetListProductsQuery) (*ListProductsResult, error)
}

type getListProductsHandler struct {
	repo Repository
}

func NewGetListProductsHandler(repo Repository) GetListProductsQueryHandler {
	return &getListProductsHandler{repo: repo}
}

func (h *getListProductsHandler) Handle(ctx context.Context, query GetListProductsQuery) (*ListProductsResult, error) {
	listQuery := ListQuery(query)

	result, err := h.repo.FindList(ctx, listQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get products list: %w", err)
	}

	return &ListProductsResult{
		Items: result.Items,
		Page:  result.Page,
		Size:  result.Size,
		Total: result.Total,
	}, nil
}
