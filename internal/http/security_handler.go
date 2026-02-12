package http //nolint:revive // package name intentional

import (
	"context"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/httpapi"
	"github.com/Sokol111/ecommerce-commons/pkg/security/token"
)

type securityHandler struct {
	handler token.SecurityHandler
}

func newSecurityHandler(handler token.SecurityHandler) httpapi.SecurityHandler {
	return &securityHandler{handler: handler}
}

// HandleBearerAuth handles BearerAuth security.
func (s *securityHandler) HandleBearerAuth(ctx context.Context, operationName httpapi.OperationName, t httpapi.BearerAuth) (context.Context, error) {
	ctx, _, err := s.handler.HandleBearerAuth(ctx, t.Token, t.Roles)
	return ctx, err
}
