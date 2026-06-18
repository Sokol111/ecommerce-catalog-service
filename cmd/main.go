package main

import (
	"context"

	"github.com/Sokol111/ecommerce-catalog-service/internal/application"
	internalconnect "github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/inbound/connect"
	"github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/outbound/kafka"
	"github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/outbound/mongo"
	commons_core "github.com/Sokol111/ecommerce-commons/pkg/core"
	commons_http "github.com/Sokol111/ecommerce-commons/pkg/http"
	commons_messaging "github.com/Sokol111/ecommerce-commons/pkg/messaging"
	commons_observability "github.com/Sokol111/ecommerce-commons/pkg/observability"
	commons_persistence "github.com/Sokol111/ecommerce-commons/pkg/persistence"
	commons_token "github.com/Sokol111/ecommerce-commons/pkg/security/token"
	commons_validation "github.com/Sokol111/ecommerce-commons/pkg/security/validation"
	"github.com/Sokol111/ecommerce-commons/pkg/tenant"
	tenant_api_client "github.com/Sokol111/ecommerce-tenant-service-api/pkg/client"
	tenant_api_consumer "github.com/Sokol111/ecommerce-tenant-service-api/pkg/consumer"
	tenant_api_provider "github.com/Sokol111/ecommerce-tenant-service-api/pkg/provider"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var AppModules = fx.Options(
	// Commons
	commons_core.NewCoreModule(),
	commons_persistence.NewPersistenceModule(),
	commons_http.NewHTTPModule(commons_http.WithH2C()),
	commons_observability.NewObservabilityModule(),
	commons_messaging.NewMessagingModule(),
	commons_validation.NewModule(),
	commons_token.NewModule(),

	// Tenant
	tenant.NewModule(tenant.WithMigrations()),
	tenant_api_consumer.Module(),
	tenant_api_provider.Module(),
	tenant_api_client.Module(),

	// Domain & Application
	mongo.Module(),
	application.Module(),
	kafka.Module(),

	// Connect (gRPC/Connect-RPC)
	internalconnect.Module(),
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
