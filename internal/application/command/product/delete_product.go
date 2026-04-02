package product

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	"github.com/Sokol111/ecommerce-catalog-service/internal/event"
	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"go.uber.org/zap"
)

type DeleteProductCommand struct {
	ID string
}

type DeleteProductCommandHandler interface {
	Handle(ctx context.Context, cmd DeleteProductCommand) error
}

type deleteProductHandler struct {
	repo         product.Repository
	outbox       outbox.Outbox
	txManager    mongo.TxManager
	eventFactory event.ProductEventFactory
}

func NewDeleteProductHandler(
	repo product.Repository,
	outbox outbox.Outbox,
	txManager mongo.TxManager,
	eventFactory event.ProductEventFactory,
) DeleteProductCommandHandler {
	return &deleteProductHandler{
		repo:         repo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *deleteProductHandler) Handle(ctx context.Context, cmd DeleteProductCommand) error {
	msg := h.eventFactory.NewProductDeletedOutboxMessage(ctx, cmd.ID)

	send, err := mongo.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (outbox.SendFunc, error) {
		if err := h.repo.Delete(txCtx, cmd.ID); err != nil {
			return nil, fmt.Errorf("failed to delete product: %w", err)
		}

		send, err := h.outbox.Create(txCtx, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to create outbox: %w", err)
		}

		return send, nil
	})
	if err != nil {
		return err
	}

	h.log(ctx).Debug("product deleted", zap.String("id", cmd.ID))

	_ = send(ctx) //nolint:errcheck // best-effort send, errors already logged in outbox

	return nil
}

func (h *deleteProductHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "delete-product-handler"))
}
