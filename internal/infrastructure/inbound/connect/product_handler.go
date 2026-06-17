package connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	catalogv1 "github.com/Sokol111/ecommerce-catalog-service-api/gen/connect/catalog/v1"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/product"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type productHandler struct {
	createHandler  product.CreateProductCommandHandler
	updateHandler  product.UpdateProductCommandHandler
	deleteHandler  product.DeleteProductCommandHandler
	getByIDHandler product.GetProductByIDQueryHandler
	getListHandler product.GetListProductsQueryHandler
}

func (h *productHandler) CreateProduct(ctx context.Context, req *connect.Request[catalogv1.CreateProductRequest]) (*connect.Response[catalogv1.CreateProductResponse], error) {
	cmd := product.CreateProductCommand{
		Name:        req.Msg.GetName(),
		Description: req.Msg.Description,
		Price:       req.Msg.GetPrice(),
		Quantity:    int(req.Msg.GetQuantity()),
		ImageID:     req.Msg.ImageId,
		CategoryID:  req.Msg.CategoryId,
		Enabled:     req.Msg.GetEnabled(),
		Attributes:  protoToAttributeValues(req.Msg.GetAttributes()),
	}
	if req.Msg.Id != nil {
		cmd.ID = parseUUIDPtr(*req.Msg.Id)
	}

	created, err := h.createHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, mapProductConnectError(err)
	}

	return connect.NewResponse(&catalogv1.CreateProductResponse{
		Product: toProtoProduct(created),
	}), nil
}

func (h *productHandler) UpdateProduct(ctx context.Context, req *connect.Request[catalogv1.UpdateProductRequest]) (*connect.Response[catalogv1.UpdateProductResponse], error) {
	cmd := product.UpdateProductCommand{
		ID:          req.Msg.GetId(),
		Version:     int(req.Msg.GetVersion()),
		Name:        req.Msg.GetName(),
		Description: req.Msg.Description,
		Price:       req.Msg.GetPrice(),
		Quantity:    int(req.Msg.GetQuantity()),
		ImageID:     req.Msg.ImageId,
		CategoryID:  req.Msg.CategoryId,
		Enabled:     req.Msg.GetEnabled(),
		Attributes:  protoToAttributeValues(req.Msg.GetAttributes()),
	}

	updated, err := h.updateHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, mapProductConnectError(err)
	}

	return connect.NewResponse(&catalogv1.UpdateProductResponse{
		Product: toProtoProduct(updated),
	}), nil
}

func (h *productHandler) GetProductById(ctx context.Context, req *connect.Request[catalogv1.GetProductByIdRequest]) (*connect.Response[catalogv1.GetProductByIdResponse], error) { //nolint:revive
	q := product.GetProductByIDQuery{ID: req.Msg.GetId()}

	found, err := h.getByIDHandler.Handle(ctx, q)
	if err != nil {
		return nil, mapProductConnectError(err)
	}

	return connect.NewResponse(&catalogv1.GetProductByIdResponse{
		Product: toProtoProduct(found),
	}), nil
}

func (h *productHandler) DeleteProduct(ctx context.Context, req *connect.Request[catalogv1.DeleteProductRequest]) (*connect.Response[catalogv1.DeleteProductResponse], error) {
	cmd := product.DeleteProductCommand{ID: req.Msg.GetId()}

	if err := h.deleteHandler.Handle(ctx, cmd); err != nil {
		return nil, mapProductConnectError(err)
	}

	return connect.NewResponse(&catalogv1.DeleteProductResponse{}), nil
}

func (h *productHandler) GetProductList(ctx context.Context, req *connect.Request[catalogv1.GetProductListRequest]) (*connect.Response[catalogv1.GetProductListResponse], error) {
	q := product.GetListProductsQuery{
		Page:       int(req.Msg.GetPage()),
		Size:       int(req.Msg.GetSize()),
		Enabled:    req.Msg.Enabled,
		CategoryID: req.Msg.CategoryId,
		Sort:       req.Msg.GetSort(),
		Order:      req.Msg.GetOrder(),
	}

	result, err := h.getListHandler.Handle(ctx, q)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	items := make([]*catalogv1.Product, len(result.Items))
	for i, p := range result.Items {
		items[i] = toProtoProduct(p)
	}

	return connect.NewResponse(&catalogv1.GetProductListResponse{
		Items: items,
		Page:  int32(result.Page), //nolint:gosec // Page originates from int32 proto field, cannot overflow
		Size:  int32(result.Size), //nolint:gosec // Size originates from int32 proto field, cannot overflow
		Total: result.Total,
	}), nil
}

// ==================== Helpers ====================

func toProtoProduct(p *product.Product) *catalogv1.Product {
	attrs := make([]*catalogv1.AttributeValue, len(p.Attributes))
	for i, a := range p.Attributes {
		attrs[i] = domainToProtoAttributeValue(a)
	}
	return &catalogv1.Product{
		Id:          p.ID,
		Version:     int32(p.Version), //nolint:gosec // Version is an optimistic lock counter, cannot realistically overflow int32
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Quantity:    int32(p.Quantity), //nolint:gosec // Quantity is a product inventory count, practically bounded
		ImageId:     p.ImageID,
		CategoryId:  p.CategoryID,
		Enabled:     p.Enabled,
		Attributes:  attrs,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		ModifiedAt:  timestamppb.New(p.ModifiedAt),
	}
}

func domainToProtoAttributeValue(a product.AttributeValue) *catalogv1.AttributeValue {
	av := &catalogv1.AttributeValue{AttributeId: a.AttributeID}
	switch {
	case a.OptionSlugValue != nil:
		av.Value = &catalogv1.AttributeValue_OptionSlugValue{OptionSlugValue: *a.OptionSlugValue}
	case len(a.OptionSlugValues) > 0:
		av.Value = &catalogv1.AttributeValue_OptionSlugValues{
			OptionSlugValues: &catalogv1.StringList{Values: a.OptionSlugValues},
		}
	case a.NumericValue != nil:
		av.Value = &catalogv1.AttributeValue_NumericValue{NumericValue: *a.NumericValue}
	case a.TextValue != nil:
		av.Value = &catalogv1.AttributeValue_TextValue{TextValue: *a.TextValue}
	case a.BooleanValue != nil:
		av.Value = &catalogv1.AttributeValue_BooleanValue{BooleanValue: *a.BooleanValue}
	}
	return av
}

func protoToAttributeValues(attrs []*catalogv1.AttributeValueInput) []product.AttributeValue {
	result := make([]product.AttributeValue, len(attrs))
	for i, a := range attrs {
		result[i] = protoToAttributeValue(a)
	}
	return result
}

func protoToAttributeValue(a *catalogv1.AttributeValueInput) product.AttributeValue {
	av := product.AttributeValue{AttributeID: a.GetAttributeId()}
	switch v := a.Value.(type) {
	case *catalogv1.AttributeValueInput_OptionSlugValue:
		av.OptionSlugValue = &v.OptionSlugValue
	case *catalogv1.AttributeValueInput_OptionSlugValues:
		if v.OptionSlugValues != nil {
			av.OptionSlugValues = v.OptionSlugValues.Values
		}
	case *catalogv1.AttributeValueInput_NumericValue:
		av.NumericValue = &v.NumericValue
	case *catalogv1.AttributeValueInput_TextValue:
		av.TextValue = &v.TextValue
	case *catalogv1.AttributeValueInput_BooleanValue:
		av.BooleanValue = &v.BooleanValue
	}
	return av
}

func mapProductConnectError(err error) *connect.Error {
	switch {
	case errors.Is(err, product.ErrInvalidProductData):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, product.ErrCategoryNotFound):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, mongo.ErrEntityNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, mongo.ErrOptimisticLocking):
		return connect.NewError(connect.CodeAborted, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
