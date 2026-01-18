package product

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
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
	Items []*product.Product
	Page  int
	Size  int
	Total int64
}

type GetListProductsQueryHandler interface {
	Handle(ctx context.Context, query GetListProductsQuery) (*ListProductsResult, error)
}

type getListProductsHandler struct {
	repo product.Repository
}

func NewGetListProductsHandler(repo product.Repository) GetListProductsQueryHandler {
	return &getListProductsHandler{repo: repo}
}

func (h *getListProductsHandler) Handle(ctx context.Context, query GetListProductsQuery) (*ListProductsResult, error) {
	listQuery := product.ListQuery{
		Page:       query.Page,
		Size:       query.Size,
		Enabled:    query.Enabled,
		CategoryID: query.CategoryID,
		Sort:       query.Sort,
		Order:      query.Order,
	}

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
