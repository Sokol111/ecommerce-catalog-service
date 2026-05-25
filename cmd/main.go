package main

import (
	"context"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/httpapi"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application"
	"github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/inbound/http"
	"github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/outbound/kafka"
	"github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/outbound/mongo"
	commons_core "github.com/Sokol111/ecommerce-commons/pkg/core"
	commons_http "github.com/Sokol111/ecommerce-commons/pkg/http"
	commons_messaging "github.com/Sokol111/ecommerce-commons/pkg/messaging"
	commons_observability "github.com/Sokol111/ecommerce-commons/pkg/observability"
	commons_persistence "github.com/Sokol111/ecommerce-commons/pkg/persistence"
	commons_pyroscope "github.com/Sokol111/ecommerce-commons/pkg/pyroscope"
	commons_token "github.com/Sokol111/ecommerce-commons/pkg/security/token"
	commons_validation "github.com/Sokol111/ecommerce-commons/pkg/security/validation"
	commons_swaggerui "github.com/Sokol111/ecommerce-commons/pkg/swaggerui"
	"github.com/Sokol111/ecommerce-commons/pkg/tenant"
	tenantapi "github.com/Sokol111/ecommerce-tenant-service-api/gen/httpapi"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var AppModules = fx.Options(
	// Commons
	commons_core.NewCoreModule(),
	commons_persistence.NewPersistenceModule(commons_persistence.WithTenantMigrations()),
	commons_http.NewHTTPModule(),
	commons_observability.NewObservabilityModule(),
	commons_messaging.NewMessagingModule(),
	commons_validation.NewModule(),
	commons_token.NewModule(),
	commons_pyroscope.NewPyroscopeModule(),
	commons_swaggerui.NewSwaggerModule(),

	// Tenant
	tenant.MiddlewareModule(),
	tenantapi.NewTenantSlugsModule("clients.tenant-service"),
	tenantapi.TenantEventsModule("tenant-events"),

	// Domain & Application
	mongo.Module(),
	application.Module(),
	kafka.Module(),

	// HTTP
	httpapi.ServerModule(),
	http.Module(),
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
