package product

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	"github.com/Sokol111/ecommerce-catalog-service/internal/event"
	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type CreateProductCommand struct {
	ID          *uuid.UUID
	Name        string
	Description *string
	Price       float32
	Quantity    int
	ImageID     *string
	CategoryID  *string
	Enabled     bool
	Attributes  []product.AttributeValue
}

type CreateProductCommandHandler interface {
	Handle(ctx context.Context, cmd CreateProductCommand) (*product.Product, error)
}

type createProductHandler struct {
	repo         product.Repository
	attrRepo     attribute.Repository
	categoryRepo category.Repository
	outbox       outbox.Outbox
	txManager    mongo.TxManager
	eventFactory event.ProductEventFactory
}

func NewCreateProductHandler(
	repo product.Repository,
	attrRepo attribute.Repository,
	categoryRepo category.Repository,
	outbox outbox.Outbox,
	txManager mongo.TxManager,
	eventFactory event.ProductEventFactory,
) CreateProductCommandHandler {
	return &createProductHandler{
		repo:         repo,
		attrRepo:     attrRepo,
		categoryRepo: categoryRepo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *createProductHandler) Handle(ctx context.Context, cmd CreateProductCommand) (*product.Product, error) {
	if err := h.validateCategory(ctx, cmd.CategoryID); err != nil {
		return nil, err
	}

	if err := h.validateAttributes(ctx, cmd.Attributes); err != nil {
		return nil, err
	}

	p, err := h.createProduct(cmd)
	if err != nil {
		return nil, err
	}

	msg := h.eventFactory.NewProductUpdatedOutboxMessage(ctx, p)

	return h.persistAndPublish(ctx, p, msg)
}

func (h *createProductHandler) validateCategory(ctx context.Context, categoryID *string) error {
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

func (h *createProductHandler) validateAttributes(ctx context.Context, productAttrs []product.AttributeValue) error {
	attrIDs := lo.Map(productAttrs, func(attr product.AttributeValue, _ int) string {
		return attr.AttributeID
	})
	_, err := h.attrRepo.FindByIDsOrFail(ctx, attrIDs)
	return err
}

func (h *createProductHandler) createProduct(cmd CreateProductCommand) (*product.Product, error) {
	var p *product.Product
	var err error

	if cmd.ID != nil {
		p, err = product.NewProductWithID(cmd.ID.String(), cmd.Name, cmd.Description, cmd.Price, cmd.Quantity, cmd.ImageID, cmd.CategoryID, cmd.Enabled, cmd.Attributes)
	} else {
		p, err = product.NewProduct(cmd.Name, cmd.Description, cmd.Price, cmd.Quantity, cmd.ImageID, cmd.CategoryID, cmd.Enabled, cmd.Attributes)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}
	return p, nil
}

func (h *createProductHandler) persistAndPublish(
	ctx context.Context,
	p *product.Product,
	msg outbox.Message,
) (*product.Product, error) {
	type createResult struct {
		Product *product.Product
		Send    outbox.SendFunc
	}

	res, err := mongo.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*createResult, error) {
		if err := h.repo.Insert(txCtx, p); err != nil {
			return nil, fmt.Errorf("failed to insert product: %w", err)
		}

		send, err := h.outbox.Create(txCtx, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to create outbox: %w", err)
		}

		return &createResult{
			Product: p,
			Send:    send,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	h.log(ctx).Debug("product created", zap.String("id", res.Product.ID))

	_ = res.Send(ctx) //nolint:errcheck // best-effort send, errors already logged in outbox

	return res.Product, nil
}

func (h *createProductHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "create-product-handler"))
}
