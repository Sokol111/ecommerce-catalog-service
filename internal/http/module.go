package http //nolint:revive // package name intentional

import (
	"net/http"

	"go.uber.org/fx"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/httpapi"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			newProductHandler,
			newCategoryHandler,
			newAttributeHandler,
			newCatalogHandler,
			httpapi.ProvideServer,
			newSecurityHandler,
		),
		fx.Invoke(registerOgenRoutes),
	)
}

type catalogHandler struct {
	*productHandler
	*categoryHandler
	*attributeHandler
}

func newCatalogHandler(productHandler *productHandler, categoryHandler *categoryHandler, attributeHandler *attributeHandler) httpapi.Handler {
	return &catalogHandler{
		productHandler:   productHandler,
		categoryHandler:  categoryHandler,
		attributeHandler: attributeHandler,
	}
}

func registerOgenRoutes(mux *http.ServeMux, server *httpapi.Server) {
	mux.Handle("/", server)
}
