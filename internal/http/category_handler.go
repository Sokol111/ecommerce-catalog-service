package http

import (
	"context"
	"errors"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/httpapi"
	command "github.com/Sokol111/ecommerce-catalog-service/internal/application/command/category"
	query "github.com/Sokol111/ecommerce-catalog-service/internal/application/query/category"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

type categoryHandler struct {
	createHandler  command.CreateCategoryCommandHandler
	updateHandler  command.UpdateCategoryCommandHandler
	getByIDHandler query.GetCategoryByIDQueryHandler
	getListHandler query.GetListCategoriesQueryHandler
}

func newCategoryHandler(
	createHandler command.CreateCategoryCommandHandler,
	updateHandler command.UpdateCategoryCommandHandler,
	getByIDHandler query.GetCategoryByIDQueryHandler,
	getListHandler query.GetListCategoriesQueryHandler,
) *categoryHandler {
	return &categoryHandler{
		createHandler:  createHandler,
		updateHandler:  updateHandler,
		getByIDHandler: getByIDHandler,
		getListHandler: getListHandler,
	}
}

func toCategoryAttributeResponse(attr category.CategoryAttribute, _ int) httpapi.CategoryAttribute {
	return httpapi.CategoryAttribute{
		AttributeId: attr.AttributeID,
		Role:        httpapi.CategoryAttributeRole(attr.Role),
		Required:    attr.Required,
		SortOrder:   attr.SortOrder,
		Filterable:  attr.Filterable,
		Searchable:  attr.Searchable,
	}
}

func toCategoryResponse(c *category.Category) *httpapi.CategoryResponse {
	return &httpapi.CategoryResponse{
		ID:         c.ID,
		Version:    c.Version,
		Name:       c.Name,
		Enabled:    c.Enabled,
		Attributes: lo.Map(c.Attributes, toCategoryAttributeResponse),
		CreatedAt:  c.CreatedAt,
		ModifiedAt: c.ModifiedAt,
	}
}

func toAttributeInput(attr httpapi.CategoryAttributeInput, _ int) command.CategoryAttributeInput {
	return command.CategoryAttributeInput{
		AttributeID: attr.AttributeId.String(),
		Role:        string(attr.Role),
		Required:    attr.Required.Or(false),
		SortOrder:   attr.SortOrder.Or(0),
		Filterable:  attr.Filterable,
		Searchable:  attr.Searchable,
	}
}

func (h *categoryHandler) CreateCategory(ctx context.Context, req *httpapi.CreateCategoryRequest) (httpapi.CreateCategoryRes, error) {
	cmd := command.CreateCategoryCommand{
		ID:         lo.If(req.ID.IsSet(), &req.ID.Value).Else(nil),
		Name:       req.Name,
		Enabled:    req.Enabled,
		Attributes: lo.Map(req.Attributes, toAttributeInput),
	}

	created, err := h.createHandler.Handle(ctx, cmd)
	if err != nil {
		if errors.Is(err, category.ErrInvalidCategoryData) {
			return &httpapi.CreateCategoryBadRequest{
				Status: 400,
				Type:   *aboutBlankURL,
				Title:  err.Error(),
			}, nil
		}
		return nil, err
	}

	return toCategoryResponse(created), nil
}

func (h *categoryHandler) GetCategoryById(ctx context.Context, params httpapi.GetCategoryByIdParams) (httpapi.GetCategoryByIdRes, error) {
	q := query.GetCategoryByIDQuery{ID: params.ID.String()}

	found, err := h.getByIDHandler.Handle(ctx, q)
	if errors.Is(err, mongo.ErrEntityNotFound) {
		return &httpapi.GetCategoryByIdNotFound{
			Status: 404,
			Type:   *aboutBlankURL,
			Title:  "Category not found",
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return toCategoryResponse(found), nil
}

func (h *categoryHandler) GetCategoryList(ctx context.Context, params httpapi.GetCategoryListParams) (httpapi.GetCategoryListRes, error) {
	q := query.GetListCategoriesQuery{
		Page:    params.Page,
		Size:    params.Size,
		Enabled: lo.ToPtr(params.Enabled.Or(false)),
		Sort:    string(params.Sort.Or(httpapi.GetCategoryListSortCreatedAt)),
		Order:   string(params.Order.Or(httpapi.GetCategoryListOrderDesc)),
	}

	if !params.Enabled.IsSet() {
		q.Enabled = nil
	}

	result, err := h.getListHandler.Handle(ctx, q)
	if err != nil {
		return nil, err
	}

	return &httpapi.CategoryListResponse{
		Items: lo.Map(result.Items, func(c *category.Category, _ int) httpapi.CategoryResponse {
			return *toCategoryResponse(c)
		}),
		Page:  result.Page,
		Size:  result.Size,
		Total: int(result.Total),
	}, nil
}

func (h *categoryHandler) UpdateCategory(ctx context.Context, req *httpapi.UpdateCategoryRequest) (httpapi.UpdateCategoryRes, error) {
	cmd := command.UpdateCategoryCommand{
		ID:         req.ID.String(),
		Version:    req.Version,
		Name:       req.Name,
		Enabled:    req.Enabled,
		Attributes: lo.Map(req.Attributes, toAttributeInput),
	}

	updated, err := h.updateHandler.Handle(ctx, cmd)
	if err != nil {
		if errors.Is(err, category.ErrInvalidCategoryData) {
			return &httpapi.UpdateCategoryBadRequest{
				Status: 400,
				Type:   *aboutBlankURL,
				Title:  err.Error(),
			}, nil
		}
		if errors.Is(err, mongo.ErrEntityNotFound) {
			return &httpapi.UpdateCategoryBadRequest{
				Status: 400,
				Type:   *aboutBlankURL,
				Title:  "Category not found",
			}, nil
		}
		if errors.Is(err, mongo.ErrOptimisticLocking) {
			return &httpapi.UpdateCategoryPreconditionFailed{
				Status: 412,
				Type:   *aboutBlankURL,
				Title:  "Version mismatch",
			}, nil
		}
		return nil, err
	}

	return toCategoryResponse(updated), nil
}
