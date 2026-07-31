package fxconfig

import (
	"net/http"

	"connectrpc.com/connect"
	catalogv1connect "github.com/Sokol111/ecommerce-catalog-service-api/gen/go/catalog/v1/catalogv1connect"
	internalconnect "github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/inbound/connect"
	"github.com/Sokol111/ecommerce-commons/pkg/security/validation"
	"go.uber.org/fx"
)

// NewConnectModule provides the Connect gRPC/Connect-RPC server handlers for catalog operations.
func NewConnectModule() fx.Option {
	return fx.Options(
		fx.Provide(
			internalconnect.NewAttributeHandler,
			internalconnect.NewCategoryHandler,
			internalconnect.NewProductHandler,
			provideProcedurePermissions,
		),
		fx.Invoke(registerConnectRoutes),
	)
}

func registerConnectRoutes(
	mux *http.ServeMux,
	attrHandler *internalconnect.AttributeHandler,
	catHandler *internalconnect.CategoryHandler,
	prodHandler *internalconnect.ProductHandler,
	interceptors []connect.Interceptor,
) {
	opts := connect.WithInterceptors(interceptors...)

	attrPath, attrH := catalogv1connect.NewAttributeServiceHandler(attrHandler, opts)
	mux.Handle(attrPath, attrH)

	catPath, catH := catalogv1connect.NewCategoryServiceHandler(catHandler, opts)
	mux.Handle(catPath, catH)

	prodPath, prodH := catalogv1connect.NewProductServiceHandler(prodHandler, opts)
	mux.Handle(prodPath, prodH)
}

func provideProcedurePermissions() validation.ProcedurePermissions {
	return validation.ProcedurePermissions{
		catalogv1connect.AttributeServiceCreateAttributeProcedure:  {"attributes:write"},
		catalogv1connect.AttributeServiceUpdateAttributeProcedure:  {"attributes:write"},
		catalogv1connect.AttributeServiceGetAttributeByIdProcedure: {"attributes:read"},
		catalogv1connect.AttributeServiceGetAttributeListProcedure: {"attributes:read"},
		catalogv1connect.CategoryServiceCreateCategoryProcedure:    {"categories:write"},
		catalogv1connect.CategoryServiceUpdateCategoryProcedure:    {"categories:write"},
		catalogv1connect.CategoryServiceGetCategoryByIdProcedure:   {"categories:read"},
		catalogv1connect.CategoryServiceGetCategoryListProcedure:   {"categories:read"},
		catalogv1connect.ProductServiceCreateProductProcedure:      {"products:write"},
		catalogv1connect.ProductServiceUpdateProductProcedure:      {"products:write"},
		catalogv1connect.ProductServiceDeleteProductProcedure:      {"products:delete"},
		catalogv1connect.ProductServiceGetProductByIdProcedure:     {"products:read"},
		catalogv1connect.ProductServiceGetProductListProcedure:     {"products:read"},
	}
}
