package http

import (
	"context"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/httpapi"
	"github.com/Sokol111/ecommerce-commons/pkg/security/token"
)

type securityHandler struct {
	tokenValidator token.Validator
}

func newSecurityHandler(tokenValidator token.Validator) httpapi.SecurityHandler {
	return &securityHandler{tokenValidator: tokenValidator}
}

// HandleBearerAuth handles BearerAuth security.
func (s *securityHandler) HandleBearerAuth(ctx context.Context, operationName httpapi.OperationName, t httpapi.BearerAuth) (context.Context, error) {
	ctx, _, err := token.HandleBearerAuth(ctx, s.tokenValidator, t.Token, t.Roles)
	return ctx, err
}
