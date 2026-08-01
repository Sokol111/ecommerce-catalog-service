package fxconfig

import (
	fx_inbound "github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/inbound/fxconfig"
	fx_outbound "github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/outbound/fxconfig"
	fx_tenant_api "github.com/Sokol111/ecommerce-tenant-service-api/pkg/fxconfig"
	"go.uber.org/fx"
)

func NewInfrastructureModule() fx.Option {
	return fx.Options(
		NewTenantInfrastructureModule(),
		fx_outbound.NewOutboundInfrastructureModule(),
		fx_inbound.NewInboundInfrastructureModule(),
	)
}

func NewTenantInfrastructureModule() fx.Option {
	return fx.Options(
		fx_tenant_api.NewKafkaConsumerModule(),
		fx_tenant_api.NewTenantSlugsProviderModule(),
		fx_tenant_api.NewGrpcClientsModule(),
	)
}
