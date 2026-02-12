package product

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	"github.com/Sokol111/ecommerce-catalog-service/internal/event"
	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
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
	categoryRepo category.Repository
	outbox       outbox.Outbox
	txManager    mongo.TxManager
	eventFactory event.ProductEventFactory
}

func NewUpdateProductHandler(
	repo product.Repository,
	attrRepo attribute.Repository,
	categoryRepo category.Repository,
	outbox outbox.Outbox,
	txManager mongo.TxManager,
	eventFactory event.ProductEventFactory,
) UpdateProductCommandHandler {
	return &updateProductHandler{
		repo:         repo,
		attrRepo:     attrRepo,
		categoryRepo: categoryRepo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *updateProductHandler) Handle(ctx context.Context, cmd UpdateProductCommand) (*product.Product, error) {
	p, err := h.findAndValidateProduct(ctx, cmd.ID, cmd.Version)
	if err != nil {
		return nil, err
	}

	if err := h.validateCategory(ctx, cmd.CategoryID); err != nil {
		return nil, err
	}

	if err := h.validateAttributes(ctx, cmd.Attributes); err != nil {
		return nil, err
	}

	if err := p.Update(cmd.Name, cmd.Description, cmd.Price, cmd.Quantity, cmd.ImageID, cmd.CategoryID, cmd.Enabled, cmd.Attributes); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return h.persistAndPublish(ctx, p)
}

func (h *updateProductHandler) findAndValidateProduct(ctx context.Context, id string, version int) (*product.Product, error) {
	p, err := h.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrEntityNotFound) {
			return nil, mongo.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if p.Version != version {
		return nil, mongo.ErrOptimisticLocking
	}

	return p, nil
}

func (h *updateProductHandler) validateCategory(ctx context.Context, categoryID *string) error {
	if categoryID == nil {
		return nil
	}

	exists, err := h.categoryRepo.Exists(ctx, *categoryID)
	if err != nil {
		return fmt.Errorf("failed to check category: %w", err)
	}
	if !exists {
		return product.ErrCategoryNotFound
	}
	return nil
}

func (h *updateProductHandler) validateAttributes(ctx context.Context, productAttrs []product.AttributeValue) error {
	attrIDs := lo.Map(productAttrs, func(attr product.AttributeValue, _ int) string {
		return attr.AttributeID
	})
	_, err := h.attrRepo.FindByIDsOrFail(ctx, attrIDs)
	return err
}

func (h *updateProductHandler) persistAndPublish(
	ctx context.Context,
	p *product.Product,
) (*product.Product, error) {
	type updateResult struct {
		Product *product.Product
		Send    outbox.SendFunc
	}

	res, err := mongo.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*updateResult, error) {
		updated, err := h.repo.Update(txCtx, p)
		if err != nil {
			if errors.Is(err, mongo.ErrOptimisticLocking) {
				return nil, err
			}
			return nil, fmt.Errorf("failed to update product: %w", err)
		}

		msg := h.eventFactory.NewProductUpdatedOutboxMessage(txCtx, updated)

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

	_ = res.Send(ctx) //nolint:errcheck // best-effort send, errors already logged in outbox

	return res.Product, nil
}

func (h *updateProductHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "update-product-handler"))
}
