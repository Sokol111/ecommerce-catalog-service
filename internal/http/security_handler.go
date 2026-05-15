package http //nolint:revive // package name intentional

import (
	"context"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/httpapi"
	"github.com/Sokol111/ecommerce-commons/pkg/security/validation"
)

type securityHandler struct {
	handler validation.SecurityHandler
}

func newSecurityHandler(handler validation.SecurityHandler) httpapi.SecurityHandler {
	return &securityHandler{handler: handler}
}

// HandleBearerAuth handles BearerAuth security.
func (s *securityHandler) HandleBearerAuth(ctx context.Context, operationName httpapi.OperationName, t httpapi.BearerAuth) (context.Context, error) {
	ctx, _, err := s.handler.HandleBearerAuth(ctx, t.Token, t.Roles)
	return ctx, err
}
