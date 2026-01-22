package product

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
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
	Attributes  []product.AttributeValue
}

type UpdateProductCommandHandler interface {
	Handle(ctx context.Context, cmd UpdateProductCommand) (*product.Product, error)
}

type updateProductHandler struct {
	repo         product.Repository
	attrRepo     attribute.Repository
	outbox       outbox.Outbox
	txManager    persistence.TxManager
	eventFactory event.ProductEventFactory
}

func NewUpdateProductHandler(
	repo product.Repository,
	attrRepo attribute.Repository,
	outbox outbox.Outbox,
	txManager persistence.TxManager,
	eventFactory event.ProductEventFactory,
) UpdateProductCommandHandler {
	return &updateProductHandler{
		repo:         repo,
		attrRepo:     attrRepo,
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

	attrIDs := lo.Map(cmd.Attributes, func(attr product.AttributeValue, _ int) string {
		return attr.AttributeID
	})
	attrs, err := h.attrRepo.FindByIDsOrFail(ctx, attrIDs)
	if err != nil {
		return nil, err
	}

	if err := p.Update(cmd.Name, cmd.Description, cmd.Price, cmd.Quantity, cmd.ImageID, cmd.CategoryID, cmd.Enabled, cmd.Attributes); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	type updateResult struct {
		Product *product.Product
		Send    outbox.SendFunc
	}

	res, err := persistence.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*updateResult, error) {
		updated, err := h.repo.Update(txCtx, p)
		if err != nil {
			if errors.Is(err, persistence.ErrOptimisticLocking) {
				return nil, persistence.ErrOptimisticLocking
			}
			return nil, fmt.Errorf("failed to update product: %w", err)
		}

		msg := h.eventFactory.NewProductUpdatedOutboxMessage(txCtx, updated, attrs)

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

	h.log(ctx).Debug("product updated", zap.String("id", res.Product.ID))

	_ = res.Send(ctx)

	return res.Product, nil
}

func (h *updateProductHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "update-product-handler"))
}
