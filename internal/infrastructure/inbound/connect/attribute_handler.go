package connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	catalogv1 "github.com/Sokol111/ecommerce-catalog-service-api/gen/connect/catalog/v1"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/attribute"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type attributeHandler struct {
	createHandler  attribute.CreateAttributeCommandHandler
	updateHandler  attribute.UpdateAttributeCommandHandler
	getByIDHandler attribute.GetAttributeByIDQueryHandler
	getListHandler attribute.GetAttributeListQueryHandler
}

func (h *attributeHandler) CreateAttribute(ctx context.Context, req *connect.Request[catalogv1.CreateAttributeRequest]) (*connect.Response[catalogv1.CreateAttributeResponse], error) {
	var id *string
	if req.Msg.Id != nil {
		id = req.Msg.Id
	}

	var unit *string
	if req.Msg.Unit != nil {
		unit = req.Msg.Unit
	}

	cmd := attribute.CreateAttributeCommand{
		Name:    req.Msg.GetName(),
		Slug:    req.Msg.GetSlug(),
		Type:    protoAttributeTypeToString(req.Msg.GetType()),
		Unit:    unit,
		Enabled: req.Msg.GetEnabled(),
		Options: protoToOptionInputs(req.Msg.GetOptions()),
	}
	if id != nil {
		cmd.ID = parseUUIDPtr(*id)
	}

	created, err := h.createHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, mapAttributeConnectError(err)
	}

	return connect.NewResponse(&catalogv1.CreateAttributeResponse{
		Attribute: toProtoAttribute(created),
	}), nil
}

func (h *attributeHandler) UpdateAttribute(ctx context.Context, req *connect.Request[catalogv1.UpdateAttributeRequest]) (*connect.Response[catalogv1.UpdateAttributeResponse], error) {
	var unit *string
	if req.Msg.Unit != nil {
		unit = req.Msg.Unit
	}

	cmd := attribute.UpdateAttributeCommand{
		ID:      req.Msg.GetId(),
		Version: int(req.Msg.GetVersion()),
		Name:    req.Msg.GetName(),
		Unit:    unit,
		Enabled: req.Msg.GetEnabled(),
		Options: protoToOptionInputs(req.Msg.GetOptions()),
	}

	updated, err := h.updateHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, mapAttributeConnectError(err)
	}

	return connect.NewResponse(&catalogv1.UpdateAttributeResponse{
		Attribute: toProtoAttribute(updated),
	}), nil
}

func (h *attributeHandler) GetAttributeById(ctx context.Context, req *connect.Request[catalogv1.GetAttributeByIdRequest]) (*connect.Response[catalogv1.GetAttributeByIdResponse], error) { //nolint:revive
	q := attribute.GetAttributeByIDQuery{ID: req.Msg.GetId()}

	found, err := h.getByIDHandler.Handle(ctx, q)
	if err != nil {
		return nil, mapAttributeConnectError(err)
	}

	return connect.NewResponse(&catalogv1.GetAttributeByIdResponse{
		Attribute: toProtoAttribute(found),
	}), nil
}

func (h *attributeHandler) GetAttributeList(ctx context.Context, req *connect.Request[catalogv1.GetAttributeListRequest]) (*connect.Response[catalogv1.GetAttributeListResponse], error) {
	var attrType *string
	if req.Msg.Type != nil {
		s := protoAttributeTypeToString(*req.Msg.Type)
		attrType = &s
	}

	q := attribute.GetAttributeListQuery{
		Page:    int(req.Msg.GetPage()),
		Size:    int(req.Msg.GetSize()),
		Enabled: req.Msg.Enabled,
		Type:    attrType,
		Sort:    req.Msg.GetSort(),
		Order:   req.Msg.GetOrder(),
	}

	result, err := h.getListHandler.Handle(ctx, q)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	items := make([]*catalogv1.Attribute, len(result.Items))
	for i, a := range result.Items {
		items[i] = toProtoAttribute(a)
	}

	return connect.NewResponse(&catalogv1.GetAttributeListResponse{
		Items: items,
		Page:  int32(result.Page), //nolint:gosec // Page originates from int32 proto field, cannot overflow
		Size:  int32(result.Size), //nolint:gosec // Size originates from int32 proto field, cannot overflow
		Total: result.Total,
	}), nil
}

// ==================== Helpers ====================

func toProtoAttribute(a *attribute.Attribute) *catalogv1.Attribute {
	opts := make([]*catalogv1.AttributeOption, len(a.Options))
	for i, o := range a.Options {
		opts[i] = &catalogv1.AttributeOption{
			Name:      o.Name,
			Slug:      o.Slug,
			ColorCode: o.ColorCode,
			SortOrder: int32(o.SortOrder), //nolint:gosec // SortOrder is a small integer, cannot overflow int32
		}
	}
	return &catalogv1.Attribute{
		Id:         a.ID,
		Version:    int32(a.Version), //nolint:gosec // Version is an optimistic lock counter, cannot realistically overflow int32
		Name:       a.Name,
		Slug:       a.Slug,
		Type:       stringToProtoAttributeType(string(a.Type)),
		Unit:       a.Unit,
		Enabled:    a.Enabled,
		Options:    opts,
		CreatedAt:  timestamppb.New(a.CreatedAt),
		ModifiedAt: timestamppb.New(a.ModifiedAt),
	}
}

func protoToOptionInputs(opts []*catalogv1.AttributeOptionInput) []attribute.OptionInput {
	result := make([]attribute.OptionInput, len(opts))
	for i, o := range opts {
		var sortOrder int
		if o.SortOrder != nil {
			sortOrder = int(*o.SortOrder)
		}
		result[i] = attribute.OptionInput{
			Name:      o.GetName(),
			Slug:      o.GetSlug(),
			ColorCode: o.ColorCode,
			SortOrder: sortOrder,
		}
	}
	return result
}

func protoAttributeTypeToString(t catalogv1.AttributeType) string {
	switch t {
	case catalogv1.AttributeType_ATTRIBUTE_TYPE_SINGLE:
		return "single"
	case catalogv1.AttributeType_ATTRIBUTE_TYPE_MULTIPLE:
		return "multiple"
	case catalogv1.AttributeType_ATTRIBUTE_TYPE_RANGE:
		return "range"
	case catalogv1.AttributeType_ATTRIBUTE_TYPE_BOOLEAN:
		return "boolean"
	case catalogv1.AttributeType_ATTRIBUTE_TYPE_TEXT:
		return "text"
	default:
		return ""
	}
}

func stringToProtoAttributeType(s string) catalogv1.AttributeType {
	switch s {
	case "single":
		return catalogv1.AttributeType_ATTRIBUTE_TYPE_SINGLE
	case "multiple":
		return catalogv1.AttributeType_ATTRIBUTE_TYPE_MULTIPLE
	case "range":
		return catalogv1.AttributeType_ATTRIBUTE_TYPE_RANGE
	case "boolean":
		return catalogv1.AttributeType_ATTRIBUTE_TYPE_BOOLEAN
	case "text":
		return catalogv1.AttributeType_ATTRIBUTE_TYPE_TEXT
	default:
		return catalogv1.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED
	}
}

func mapAttributeConnectError(err error) *connect.Error {
	switch {
	case errors.Is(err, attribute.ErrInvalidAttributeData):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, attribute.ErrSlugAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, mongo.ErrEntityNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, mongo.ErrOptimisticLocking):
		return connect.NewError(connect.CodeAborted, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
