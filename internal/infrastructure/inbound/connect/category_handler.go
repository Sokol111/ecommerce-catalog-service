package connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	catalogv1 "github.com/Sokol111/ecommerce-catalog-service-api/gen/connect/catalog/v1"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/category"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type categoryHandler struct {
	createHandler  category.CreateCategoryCommandHandler
	updateHandler  category.UpdateCategoryCommandHandler
	getByIDHandler category.GetCategoryByIDQueryHandler
	getListHandler category.GetListCategoriesQueryHandler
}

func (h *categoryHandler) CreateCategory(ctx context.Context, req *connect.Request[catalogv1.CreateCategoryRequest]) (*connect.Response[catalogv1.CreateCategoryResponse], error) {
	cmd := category.CreateCategoryCommand{
		Name:       req.Msg.GetName(),
		Enabled:    req.Msg.GetEnabled(),
		Attributes: protoToCategoryAttributeInputs(req.Msg.GetAttributes()),
	}
	if req.Msg.Id != nil {
		cmd.ID = parseUUIDPtr(*req.Msg.Id)
	}

	created, err := h.createHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, mapCategoryConnectError(err)
	}

	return connect.NewResponse(&catalogv1.CreateCategoryResponse{
		Category: toProtoCategory(created),
	}), nil
}

func (h *categoryHandler) UpdateCategory(ctx context.Context, req *connect.Request[catalogv1.UpdateCategoryRequest]) (*connect.Response[catalogv1.UpdateCategoryResponse], error) {
	cmd := category.UpdateCategoryCommand{
		ID:         req.Msg.GetId(),
		Version:    int(req.Msg.GetVersion()),
		Name:       req.Msg.GetName(),
		Enabled:    req.Msg.GetEnabled(),
		Attributes: protoToCategoryAttributeInputs(req.Msg.GetAttributes()),
	}

	updated, err := h.updateHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, mapCategoryConnectError(err)
	}

	return connect.NewResponse(&catalogv1.UpdateCategoryResponse{
		Category: toProtoCategory(updated),
	}), nil
}

func (h *categoryHandler) GetCategoryById(ctx context.Context, req *connect.Request[catalogv1.GetCategoryByIdRequest]) (*connect.Response[catalogv1.GetCategoryByIdResponse], error) { //nolint:revive
	q := category.GetCategoryByIDQuery{ID: req.Msg.GetId()}

	found, err := h.getByIDHandler.Handle(ctx, q)
	if err != nil {
		return nil, mapCategoryConnectError(err)
	}

	return connect.NewResponse(&catalogv1.GetCategoryByIdResponse{
		Category: toProtoCategory(found),
	}), nil
}

func (h *categoryHandler) GetCategoryList(ctx context.Context, req *connect.Request[catalogv1.GetCategoryListRequest]) (*connect.Response[catalogv1.GetCategoryListResponse], error) {
	q := category.GetListCategoriesQuery{
		Page:    int(req.Msg.GetPage()),
		Size:    int(req.Msg.GetSize()),
		Enabled: req.Msg.Enabled,
		Sort:    req.Msg.GetSort(),
		Order:   req.Msg.GetOrder(),
	}

	result, err := h.getListHandler.Handle(ctx, q)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	items := make([]*catalogv1.Category, len(result.Items))
	for i, c := range result.Items {
		items[i] = toProtoCategory(c)
	}

	return connect.NewResponse(&catalogv1.GetCategoryListResponse{
		Items: items,
		Page:  int32(result.Page), //nolint:gosec // Page originates from int32 proto field, cannot overflow
		Size:  int32(result.Size), //nolint:gosec // Size originates from int32 proto field, cannot overflow
		Total: result.Total,
	}), nil
}

// ==================== Helpers ====================

func toProtoCategory(c *category.Category) *catalogv1.Category {
	attrs := make([]*catalogv1.CategoryAttribute, len(c.Attributes))
	for i, a := range c.Attributes {
		attrs[i] = &catalogv1.CategoryAttribute{
			AttributeId: a.AttributeID,
			Role:        stringToProtoCategoryAttributeRole(string(a.Role)),
			SortOrder:   int32(a.SortOrder), //nolint:gosec // SortOrder is a small integer, cannot overflow int32
			Filterable:  a.Filterable,
			Searchable:  a.Searchable,
		}
	}
	return &catalogv1.Category{
		Id:         c.ID,
		Version:    int64(c.Version),
		Name:       c.Name,
		Enabled:    c.Enabled,
		Attributes: attrs,
		CreatedAt:  timestamppb.New(c.CreatedAt),
		ModifiedAt: timestamppb.New(c.ModifiedAt),
	}
}

func protoToCategoryAttributeInputs(attrs []*catalogv1.CategoryAttributeInput) []category.CategoryAttributeInput {
	result := make([]category.CategoryAttributeInput, len(attrs))
	for i, a := range attrs {
		var sortOrder int
		if a.SortOrder != nil {
			sortOrder = int(*a.SortOrder)
		}
		result[i] = category.CategoryAttributeInput{
			AttributeID: a.GetAttributeId(),
			Role:        protoCategoryAttributeRoleToString(a.GetRole()),
			SortOrder:   sortOrder,
			Filterable:  a.GetFilterable(),
			Searchable:  a.GetSearchable(),
		}
	}
	return result
}

func protoCategoryAttributeRoleToString(r catalogv1.CategoryAttributeRole) string {
	switch r {
	case catalogv1.CategoryAttributeRole_CATEGORY_ATTRIBUTE_ROLE_VARIANT:
		return "variant"
	case catalogv1.CategoryAttributeRole_CATEGORY_ATTRIBUTE_ROLE_SPECIFICATION:
		return "specification"
	default:
		return ""
	}
}

func stringToProtoCategoryAttributeRole(s string) catalogv1.CategoryAttributeRole {
	switch s {
	case "variant":
		return catalogv1.CategoryAttributeRole_CATEGORY_ATTRIBUTE_ROLE_VARIANT
	case "specification":
		return catalogv1.CategoryAttributeRole_CATEGORY_ATTRIBUTE_ROLE_SPECIFICATION
	default:
		return catalogv1.CategoryAttributeRole_CATEGORY_ATTRIBUTE_ROLE_UNSPECIFIED
	}
}

func mapCategoryConnectError(err error) *connect.Error {
	switch {
	case errors.Is(err, category.ErrInvalidCategoryData):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, mongo.ErrEntityNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, mongo.ErrOptimisticLocking):
		return connect.NewError(connect.CodeAborted, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
