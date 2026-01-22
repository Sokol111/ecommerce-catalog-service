package product

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	"github.com/Sokol111/ecommerce-catalog-service/internal/event"
	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
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
	outbox       outbox.Outbox
	txManager    persistence.TxManager
	eventFactory event.ProductEventFactory
}

func NewCreateProductHandler(
	repo product.Repository,
	attrRepo attribute.Repository,
	outbox outbox.Outbox,
	txManager persistence.TxManager,
	eventFactory event.ProductEventFactory,
) CreateProductCommandHandler {
	return &createProductHandler{
		repo:         repo,
		attrRepo:     attrRepo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *createProductHandler) Handle(ctx context.Context, cmd CreateProductCommand) (*product.Product, error) {
	attrIDs := lo.Map(cmd.Attributes, func(attr product.AttributeValue, _ int) string {
		return attr.AttributeID
	})
	attrs, err := h.attrRepo.FindByIDsOrFail(ctx, attrIDs)
	if err != nil {
		return nil, err
	}

	var p *product.Product
	if cmd.ID != nil {
		p, err = product.NewProductWithID(cmd.ID.String(), cmd.Name, cmd.Description, cmd.Price, cmd.Quantity, cmd.ImageID, cmd.CategoryID, cmd.Enabled, cmd.Attributes)
	} else {
		p, err = product.NewProduct(cmd.Name, cmd.Description, cmd.Price, cmd.Quantity, cmd.ImageID, cmd.CategoryID, cmd.Enabled, cmd.Attributes)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	msg := h.eventFactory.NewProductCreatedOutboxMessage(ctx, p, attrs)

	return h.persistAndPublish(ctx, p, msg)
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

	res, err := persistence.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*createResult, error) {
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

	_ = res.Send(ctx)

	return res.Product, nil
}

func (h *createProductHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "create-product-handler"))
}
