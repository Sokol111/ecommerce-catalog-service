package fxconfig

import (
	fx_connect "github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/inbound/connect/fxconfig"
	"go.uber.org/fx"
)

func NewInboundInfrastructureModule() fx.Option {
	return fx.Options(
		fx_connect.NewConnectModule(),
	)
}
