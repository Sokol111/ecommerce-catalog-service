package http

import (
	"context"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/httpapi"
	"github.com/Sokol111/ecommerce-commons/pkg/security/token"
)

type securityHandler struct {
	tokenValidator token.TokenValidator
}

func newSecurityHandler(tokenValidator token.TokenValidator) httpapi.SecurityHandler {
	return &securityHandler{tokenValidator: tokenValidator}
}

// HandleBearerAuth handles BearerAuth security.
func (s *securityHandler) HandleBearerAuth(ctx context.Context, operationName httpapi.OperationName, t httpapi.BearerAuth) (context.Context, error) {
	ctx, _, err := token.HandleBearerAuth(s.tokenValidator, ctx, t.Token, t.Roles)
	return ctx, err
}
