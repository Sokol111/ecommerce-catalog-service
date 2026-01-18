package product

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	"github.com/Sokol111/ecommerce-catalog-service/internal/event"
	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
	"go.uber.org/zap"
)

type UpdateProductCommand struct {
	ID          string
	Version     int
	Name        string
	Description *string
	Price       float32
	Quantity    int
	ImageID     *string
	CategoryID  *string
	Enabled     bool
	Attributes  []product.ProductAttribute
}

type UpdateProductCommandHandler interface {
	Handle(ctx context.Context, cmd UpdateProductCommand) (*product.Product, error)
}

type updateProductHandler struct {
	repo         product.Repository
	outbox       outbox.Outbox
	txManager    persistence.TxManager
	eventFactory event.ProductEventFactory
}

func NewUpdateProductHandler(repo product.Repository, outbox outbox.Outbox, txManager persistence.TxManager, eventFactory event.ProductEventFactory) UpdateProductCommandHandler {
	return &updateProductHandler{
		repo:         repo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *updateProductHandler) Handle(ctx context.Context, cmd UpdateProductCommand) (*product.Product, error) {
	p, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		if errors.Is(err, persistence.ErrEntityNotFound) {
			return nil, persistence.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if p.Version != cmd.Version {
		return nil, persistence.ErrOptimisticLocking
	}

	if err := p.Update(cmd.Name, cmd.Description, cmd.Price, cmd.Quantity, cmd.ImageID, cmd.CategoryID, cmd.Enabled, cmd.Attributes); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	type updateResult struct {
		Product *product.Product
		Send    outbox.SendFunc
	}

	result, err := h.txManager.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		// Update in repository (with optimistic locking)
		updated, err := h.repo.Update(txCtx, p)
		if err != nil {
			if errors.Is(err, persistence.ErrOptimisticLocking) {
				return nil, persistence.ErrOptimisticLocking
			}
			return nil, fmt.Errorf("failed to update product: %w", err)
		}

		msg, err := h.eventFactory.NewProductUpdatedOutboxMessage(txCtx, updated)
		if err != nil {
			return nil, err
		}

		send, err := h.outbox.Create(txCtx, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to create outbox: %w", err)
		}

		return &updateResult{
			Product: updated,
			Send:    send,
		}, nil
	})

	if err != nil {
		return nil, err
	}

	res, ok := result.(*updateResult)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	h.log(ctx).Debug("product updated", zap.String("id", res.Product.ID))

	_ = res.Send(ctx)

	return res.Product, nil
}

func (h *updateProductHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "update-product-handler"))
}
