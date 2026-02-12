package http //nolint:revive // package name intentional

import (
	"net/http"

	"github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
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
			newOgenServer,
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

func newOgenServer(
	handler httpapi.Handler,
	securityHandler httpapi.SecurityHandler,
	tracerProvider trace.TracerProvider,
	meterProvider metric.MeterProvider,
	middlewares []middleware.Middleware,
	errorHandler ogenerrors.ErrorHandler,
) (*httpapi.Server, error) {
	return httpapi.NewServer(
		handler,
		securityHandler,
		httpapi.WithTracerProvider(tracerProvider),
		httpapi.WithMeterProvider(meterProvider),
		httpapi.WithErrorHandler(errorHandler),
		httpapi.WithMiddleware(middlewares...),
	)
}

func registerOgenRoutes(mux *http.ServeMux, server *httpapi.Server) {
	mux.Handle("/", server)
}
