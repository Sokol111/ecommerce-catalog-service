package connect

import (
	"net/http"

	"connectrpc.com/connect"
	catalogv1connect "github.com/Sokol111/ecommerce-catalog-service-api/gen/connect/catalog/v1/catalogv1connect"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/category"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/product"
	"github.com/Sokol111/ecommerce-commons/pkg/security/validation"
	"go.uber.org/fx"
)

// Module provides the Connect gRPC/Connect-RPC server handlers for catalog operations.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			newAttributeHandler,
			newCategoryHandler,
			newProductHandler,
			provideProcedurePermissions,
		),
		fx.Invoke(registerConnectRoutes),
	)
}

func newAttributeHandler(
	createHandler attribute.CreateAttributeCommandHandler,
	updateHandler attribute.UpdateAttributeCommandHandler,
	getByIDHandler attribute.GetAttributeByIDQueryHandler,
	getListHandler attribute.GetAttributeListQueryHandler,
) *attributeHandler {
	return &attributeHandler{
		createHandler:  createHandler,
		updateHandler:  updateHandler,
		getByIDHandler: getByIDHandler,
		getListHandler: getListHandler,
	}
}

func newCategoryHandler(
	createHandler category.CreateCategoryCommandHandler,
	updateHandler category.UpdateCategoryCommandHandler,
	getByIDHandler category.GetCategoryByIDQueryHandler,
	getListHandler category.GetListCategoriesQueryHandler,
) *categoryHandler {
	return &categoryHandler{
		createHandler:  createHandler,
		updateHandler:  updateHandler,
		getByIDHandler: getByIDHandler,
		getListHandler: getListHandler,
	}
}

func newProductHandler(
	createHandler product.CreateProductCommandHandler,
	updateHandler product.UpdateProductCommandHandler,
	deleteHandler product.DeleteProductCommandHandler,
	getByIDHandler product.GetProductByIDQueryHandler,
	getListHandler product.GetListProductsQueryHandler,
) *productHandler {
	return &productHandler{
		createHandler:  createHandler,
		updateHandler:  updateHandler,
		deleteHandler:  deleteHandler,
		getByIDHandler: getByIDHandler,
		getListHandler: getListHandler,
	}
}

func registerConnectRoutes(
	mux *http.ServeMux,
	attrHandler *attributeHandler,
	catHandler *categoryHandler,
	prodHandler *productHandler,
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
		catalogv1connect.AttributeServiceCreateAttributeProcedure:  {"attribute:write"},
		catalogv1connect.AttributeServiceUpdateAttributeProcedure:  {"attribute:write"},
		catalogv1connect.AttributeServiceGetAttributeByIdProcedure: {"attribute:read"},
		catalogv1connect.AttributeServiceGetAttributeListProcedure: {"attribute:read"},
		catalogv1connect.CategoryServiceCreateCategoryProcedure:    {"category:write"},
		catalogv1connect.CategoryServiceUpdateCategoryProcedure:    {"category:write"},
		catalogv1connect.CategoryServiceGetCategoryByIdProcedure:   {"category:read"},
		catalogv1connect.CategoryServiceGetCategoryListProcedure:   {"category:read"},
		catalogv1connect.ProductServiceCreateProductProcedure:      {"product:write"},
		catalogv1connect.ProductServiceUpdateProductProcedure:      {"product:write"},
		catalogv1connect.ProductServiceDeleteProductProcedure:      {"product:write"},
		catalogv1connect.ProductServiceGetProductByIdProcedure:     {"product:read"},
		catalogv1connect.ProductServiceGetProductListProcedure:     {"product:read"},
	}
}
