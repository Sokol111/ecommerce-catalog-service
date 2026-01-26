package http

import (
	"context"
	"errors"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/httpapi"
	command "github.com/Sokol111/ecommerce-catalog-service/internal/application/command/attribute"
	query "github.com/Sokol111/ecommerce-catalog-service/internal/application/query/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
)

type attributeHandler struct {
	createHandler  command.CreateAttributeCommandHandler
	updateHandler  command.UpdateAttributeCommandHandler
	getByIDHandler query.GetAttributeByIDQueryHandler
	getListHandler query.GetAttributeListQueryHandler
}

func newAttributeHandler(
	createHandler command.CreateAttributeCommandHandler,
	updateHandler command.UpdateAttributeCommandHandler,
	getByIDHandler query.GetAttributeByIDQueryHandler,
	getListHandler query.GetAttributeListQueryHandler,
) *attributeHandler {
	return &attributeHandler{
		createHandler:  createHandler,
		updateHandler:  updateHandler,
		getByIDHandler: getByIDHandler,
		getListHandler: getListHandler,
	}
}

func toAttributeOptionResponse(opt attribute.Option, _ int) httpapi.AttributeOption {
	return httpapi.AttributeOption{
		Name:      opt.Name,
		Slug:      opt.Slug,
		ColorCode: toOptString(opt.ColorCode),
		SortOrder: opt.SortOrder,
	}
}

func toAttributeResponse(a *attribute.Attribute) *httpapi.AttributeResponse {
	return &httpapi.AttributeResponse{
		ID:         a.ID,
		Version:    a.Version,
		Name:       a.Name,
		Slug:       a.Slug,
		Type:       httpapi.AttributeResponseType(a.Type),
		Unit:       toOptString(a.Unit),
		Enabled:    a.Enabled,
		Options:    lo.Map(a.Options, toAttributeOptionResponse),
		CreatedAt:  a.CreatedAt,
		ModifiedAt: a.ModifiedAt,
	}
}

func toOptionInput(opt httpapi.AttributeOptionInput, _ int) command.OptionInput {
	return command.OptionInput{
		Name:      opt.Name,
		Slug:      opt.Slug,
		ColorCode: lo.If(opt.ColorCode.IsSet(), &opt.ColorCode.Value).Else(nil),
		SortOrder: opt.SortOrder.Or(0),
	}
}

func (h *attributeHandler) CreateAttribute(ctx context.Context, req *httpapi.CreateAttributeRequest) (httpapi.CreateAttributeRes, error) {
	cmd := command.CreateAttributeCommand{
		ID:      lo.If(req.ID.IsSet(), &req.ID.Value).Else(nil),
		Name:    req.Name,
		Slug:    req.Slug,
		Type:    string(req.Type),
		Unit:    lo.If(req.Unit.IsSet(), &req.Unit.Value).Else(nil),
		Enabled: req.Enabled,
		Options: lo.Map(req.Options, toOptionInput),
	}

	created, err := h.createHandler.Handle(ctx, cmd)
	if err != nil {
		if errors.Is(err, attribute.ErrInvalidAttributeData) {
			return &httpapi.CreateAttributeBadRequest{
				Status: 400,
				Type:   *aboutBlankURL,
				Title:  err.Error(),
			}, nil
		}
		if errors.Is(err, attribute.ErrSlugAlreadyExists) {
			return &httpapi.CreateAttributeConflict{
				Status: 409,
				Type:   *aboutBlankURL,
				Title:  "Attribute with this slug already exists",
			}, nil
		}
		return nil, err
	}

	return toAttributeResponse(created), nil
}

func (h *attributeHandler) GetAttributeById(ctx context.Context, params httpapi.GetAttributeByIdParams) (httpapi.GetAttributeByIdRes, error) {
	q := query.GetAttributeByIDQuery{ID: params.ID.String()}

	found, err := h.getByIDHandler.Handle(ctx, q)
	if errors.Is(err, persistence.ErrEntityNotFound) {
		return &httpapi.GetAttributeByIdNotFound{
			Status: 404,
			Type:   *aboutBlankURL,
			Title:  "Attribute not found",
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return toAttributeResponse(found), nil
}

func (h *attributeHandler) GetAttributeList(ctx context.Context, params httpapi.GetAttributeListParams) (httpapi.GetAttributeListRes, error) {
	var attrType *string
	if params.Type.IsSet() {
		t := string(params.Type.Value)
		attrType = &t
	}

	var enabled *bool
	if params.Enabled.IsSet() {
		enabled = &params.Enabled.Value
	}

	q := query.GetAttributeListQuery{
		Page:    params.Page,
		Size:    params.Size,
		Enabled: enabled,
		Type:    attrType,
		Sort:    string(params.Sort.Or(httpapi.GetAttributeListSortName)),
		Order:   string(params.Order.Or(httpapi.GetAttributeListOrderAsc)),
	}

	result, err := h.getListHandler.Handle(ctx, q)
	if err != nil {
		return nil, err
	}

	return &httpapi.AttributeListResponse{
		Items: lo.Map(result.Items, func(a *attribute.Attribute, _ int) httpapi.AttributeResponse {
			return *toAttributeResponse(a)
		}),
		Page:  result.Page,
		Size:  result.Size,
		Total: int(result.Total),
	}, nil
}

func (h *attributeHandler) UpdateAttribute(ctx context.Context, req *httpapi.UpdateAttributeRequest) (httpapi.UpdateAttributeRes, error) {
	cmd := command.UpdateAttributeCommand{
		ID:      req.ID.String(),
		Version: req.Version,
		Name:    req.Name,
		Unit:    lo.If(req.Unit.IsSet(), &req.Unit.Value).Else(nil),
		Enabled: req.Enabled,
		Options: lo.Map(req.Options, toOptionInput),
	}

	updated, err := h.updateHandler.Handle(ctx, cmd)
	if err != nil {
		if errors.Is(err, attribute.ErrInvalidAttributeData) {
			return &httpapi.UpdateAttributeBadRequest{
				Status: 400,
				Type:   *aboutBlankURL,
				Title:  err.Error(),
			}, nil
		}
		if errors.Is(err, persistence.ErrEntityNotFound) {
			return &httpapi.UpdateAttributeNotFound{
				Status: 404,
				Type:   *aboutBlankURL,
				Title:  "Attribute not found",
			}, nil
		}
		if errors.Is(err, persistence.ErrOptimisticLocking) {
			return &httpapi.UpdateAttributePreconditionFailed{
				Status: 412,
				Type:   *aboutBlankURL,
				Title:  "Version mismatch",
			}, nil
		}
		return nil, err
	}

	return toAttributeResponse(updated), nil
}

func (h *attributeHandler) DeleteAttribute(ctx context.Context, params httpapi.DeleteAttributeParams) (httpapi.DeleteAttributeRes, error) {
	return &httpapi.DeleteAttributeInternalServerError{
		Status: 500,
		Type:   *aboutBlankURL,
		Title:  "Not implemented",
	}, nil
}
