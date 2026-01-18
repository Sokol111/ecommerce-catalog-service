package main

import (
	"context"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/httpapi"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application"
	"github.com/Sokol111/ecommerce-catalog-service/internal/event"
	"github.com/Sokol111/ecommerce-catalog-service/internal/http"
	"github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/messaging/kafka"
	"github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/persistence/mongo"
	"github.com/Sokol111/ecommerce-commons/pkg/modules"
	"github.com/Sokol111/ecommerce-commons/pkg/swaggerui"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var AppModules = fx.Options(
	// Infrastructure
	modules.NewCoreModule(),
	modules.NewPersistenceModule(),
	modules.NewHTTPModule(),
	modules.NewObservabilityModule(),
	modules.NewMessagingModule(),
	// Domain & Application
	mongo.Module(),
	event.EventModule(),
	application.Module(),
	kafka.Module(),

	// HTTP
	http.NewHttpHandlerModule(),
	swaggerui.NewSwaggerModule(swaggerui.SwaggerConfig{OpenAPIContent: httpapi.OpenAPIDoc}),
)

func main() {
	app := fx.New(
		AppModules,
		fx.Invoke(func(lc fx.Lifecycle, log *zap.Logger) {
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					log.Info("Application stopping...")
					return nil
				},
			})
		}),
	)
	app.Run()
}
