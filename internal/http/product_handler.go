package http

import (
	"context"
	"errors"
	"net/url"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/httpapi"
	command "github.com/Sokol111/ecommerce-catalog-service/internal/application/command/product"
	query "github.com/Sokol111/ecommerce-catalog-service/internal/application/query/product"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

type productHandler struct {
	createHandler  command.CreateProductCommandHandler
	updateHandler  command.UpdateProductCommandHandler
	getByIDHandler query.GetProductByIDQueryHandler
	getListHandler query.GetListProductsQueryHandler
}

func newProductHandler(
	createHandler command.CreateProductCommandHandler,
	updateHandler command.UpdateProductCommandHandler,
	getByIDHandler query.GetProductByIDQueryHandler,
	getListHandler query.GetListProductsQueryHandler,
) *productHandler {
	return &productHandler{
		createHandler:  createHandler,
		updateHandler:  updateHandler,
		getByIDHandler: getByIDHandler,
		getListHandler: getListHandler,
	}
}

var aboutBlankURL, _ = url.Parse("about:blank")

func optUUIDToStringPtr(o httpapi.OptUUID) *string {
	if !o.IsSet() {
		return nil
	}
	s := o.Value.String()
	return &s
}

func toOptString(s *string) httpapi.OptString {
	if s == nil {
		return httpapi.OptString{}
	}
	return httpapi.NewOptString(*s)
}

func toOptFloat64(f *float32) httpapi.OptFloat64 {
	if f == nil {
		return httpapi.OptFloat64{}
	}
	return httpapi.NewOptFloat64(float64(*f))
}

func toOptBool(b *bool) httpapi.OptBool {
	if b == nil {
		return httpapi.OptBool{}
	}
	return httpapi.NewOptBool(*b)
}

func toAttributeValueInput(attr httpapi.AttributeValueInput, _ int) product.AttributeValue {
	var numericValue *float32
	if attr.NumericValue.IsSet() {
		v := float32(attr.NumericValue.Value)
		numericValue = &v
	}

	return product.AttributeValue{
		AttributeID:      attr.AttributeId.String(),
		OptionSlugValue:  lo.If(attr.OptionSlugValue.IsSet(), &attr.OptionSlugValue.Value).Else(nil),
		OptionSlugValues: attr.OptionSlugValues,
		NumericValue:     numericValue,
		TextValue:        lo.If(attr.TextValue.IsSet(), &attr.TextValue.Value).Else(nil),
		BooleanValue:     lo.If(attr.BooleanValue.IsSet(), &attr.BooleanValue.Value).Else(nil),
	}
}

func toAttributeValueResponse(attr product.AttributeValue, _ int) httpapi.AttributeValue {
	return httpapi.AttributeValue{
		AttributeId:      attr.AttributeID,
		OptionSlugValue:  toOptString(attr.OptionSlugValue),
		OptionSlugValues: attr.OptionSlugValues,
		NumericValue:     toOptFloat64(attr.NumericValue),
		TextValue:        toOptString(attr.TextValue),
		BooleanValue:     toOptBool(attr.BooleanValue),
	}
}

func toProductResponse(p *product.Product) *httpapi.ProductResponse {
	return &httpapi.ProductResponse{
		ID:          p.ID,
		Version:     p.Version,
		Name:        p.Name,
		Description: toOptString(p.Description),
		Price:       float64(p.Price),
		Quantity:    p.Quantity,
		ImageId:     toOptString(p.ImageID),
		CategoryId:  toOptString(p.CategoryID),
		Enabled:     p.Enabled,
		Attributes:  lo.Map(p.Attributes, toAttributeValueResponse),
		CreatedAt:   p.CreatedAt,
		ModifiedAt:  p.ModifiedAt,
	}
}

func (h *productHandler) CreateProduct(ctx context.Context, req *httpapi.CreateProductRequest) (httpapi.CreateProductRes, error) {
	cmd := command.CreateProductCommand{
		ID:          lo.If(req.ID.IsSet(), &req.ID.Value).Else(nil),
		Name:        req.Name,
		Description: lo.If(req.Description.IsSet(), &req.Description.Value).Else(nil),
		Quantity:    req.Quantity,
		Price:       float32(req.Price),
		ImageID:     optUUIDToStringPtr(req.ImageId),
		CategoryID:  optUUIDToStringPtr(req.CategoryId),
		Enabled:     req.Enabled,
		Attributes:  lo.Map(req.Attributes, toAttributeValueInput),
	}

	created, err := h.createHandler.Handle(ctx, cmd)
	if err != nil {
		if errors.Is(err, product.ErrInvalidProductData) {
			return &httpapi.CreateProductBadRequest{
				Status: 400,
				Type:   *aboutBlankURL,
				Title:  err.Error(),
			}, nil
		}
		if errors.Is(err, product.ErrCategoryNotFound) {
			return &httpapi.CreateProductBadRequest{
				Status: 400,
				Type:   *aboutBlankURL,
				Title:  "Category not found",
			}, nil
		}
		return nil, err
	}

	return toProductResponse(created), nil
}

func (h *productHandler) GetProductById(ctx context.Context, params httpapi.GetProductByIdParams) (httpapi.GetProductByIdRes, error) {
	q := query.GetProductByIDQuery{ID: params.ID.String()}

	found, err := h.getByIDHandler.Handle(ctx, q)
	if errors.Is(err, mongo.ErrEntityNotFound) {
		return &httpapi.GetProductByIdNotFound{
			Status: 404,
			Type:   *aboutBlankURL,
			Title:  "Product not found",
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return toProductResponse(found), nil
}

func (h *productHandler) GetProductList(ctx context.Context, params httpapi.GetProductListParams) (httpapi.GetProductListRes, error) {
	var enabled *bool
	if params.Enabled.IsSet() {
		enabled = &params.Enabled.Value
	}

	q := query.GetListProductsQuery{
		Page:       params.Page,
		Size:       params.Size,
		Enabled:    enabled,
		CategoryID: optUUIDToStringPtr(params.CategoryId),
		Sort:       string(params.Sort.Or(httpapi.GetProductListSortCreatedAt)),
		Order:      string(params.Order.Or(httpapi.GetProductListOrderDesc)),
	}

	result, err := h.getListHandler.Handle(ctx, q)
	if err != nil {
		return nil, err
	}

	return &httpapi.ProductListResponse{
		Items: lo.Map(result.Items, func(p *product.Product, _ int) httpapi.ProductResponse {
			return *toProductResponse(p)
		}),
		Page:  result.Page,
		Size:  result.Size,
		Total: int(result.Total),
	}, nil
}

func (h *productHandler) UpdateProduct(ctx context.Context, req *httpapi.UpdateProductRequest) (httpapi.UpdateProductRes, error) {
	cmd := command.UpdateProductCommand{
		ID:          req.ID.String(),
		Version:     req.Version,
		Name:        req.Name,
		Description: lo.If(req.Description.IsSet(), &req.Description.Value).Else(nil),
		Price:       float32(req.Price),
		Quantity:    req.Quantity,
		ImageID:     optUUIDToStringPtr(req.ImageId),
		CategoryID:  optUUIDToStringPtr(req.CategoryId),
		Enabled:     req.Enabled,
		Attributes:  lo.Map(req.Attributes, toAttributeValueInput),
	}

	updated, err := h.updateHandler.Handle(ctx, cmd)
	if err != nil {
		if errors.Is(err, product.ErrInvalidProductData) {
			return &httpapi.UpdateProductBadRequest{
				Status: 400,
				Type:   *aboutBlankURL,
				Title:  err.Error(),
			}, nil
		}
		if errors.Is(err, mongo.ErrEntityNotFound) {
			return &httpapi.UpdateProductBadRequest{
				Status: 400,
				Type:   *aboutBlankURL,
				Title:  "Product not found",
			}, nil
		}
		if errors.Is(err, mongo.ErrOptimisticLocking) {
			return &httpapi.UpdateProductPreconditionFailed{
				Status: 412,
				Type:   *aboutBlankURL,
				Title:  "Version mismatch",
			}, nil
		}
		if errors.Is(err, product.ErrCategoryNotFound) {
			return &httpapi.UpdateProductBadRequest{
				Status: 400,
				Type:   *aboutBlankURL,
				Title:  "Category not found",
			}, nil
		}
		return nil, err
	}

	return toProductResponse(updated), nil
}
