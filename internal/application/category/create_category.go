package category

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service/internal/application/attribute"
	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CategoryAttributeInput represents the input for a category attribute
type CategoryAttributeInput struct {
	AttributeID string
	Role        string
	SortOrder   int
	Filterable  bool
	Searchable  bool
}

// CreateCategoryCommand represents the input for creating a category
type CreateCategoryCommand struct {
	ID         *uuid.UUID
	Name       string
	Enabled    bool
	Attributes []CategoryAttributeInput
}

// CreateCategoryCommandHandler defines the interface for creating categories
type CreateCategoryCommandHandler interface {
	Handle(ctx context.Context, cmd CreateCategoryCommand) (*Category, error)
}

type createCategoryHandler struct {
	repo         Repository
	attrRepo     attribute.Repository
	outbox       outbox.Outbox
	txManager    mongo.TxManager
	eventFactory CategoryEventFactory
}

func NewCreateCategoryHandler(
	repo Repository,
	attrRepo attribute.Repository,
	outbox outbox.Outbox,
	txManager mongo.TxManager,
	eventFactory CategoryEventFactory,
) CreateCategoryCommandHandler {
	return &createCategoryHandler{
		repo:         repo,
		attrRepo:     attrRepo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *createCategoryHandler) Handle(ctx context.Context, cmd CreateCategoryCommand) (*Category, error) {
	categoryAttrs, err := h.buildCategoryAttributes(ctx, cmd.Attributes)
	if err != nil {
		return nil, err
	}

	c, err := h.createCategory(cmd, categoryAttrs)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	msg := h.eventFactory.NewCategoryUpdatedOutboxMessage(ctx, c)

	return h.persistAndPublish(ctx, c, msg)
}

func (h *createCategoryHandler) buildCategoryAttributes(ctx context.Context, inputs []CategoryAttributeInput) ([]CategoryAttribute, error) {
	attrIDs := lo.Map(inputs, func(attr CategoryAttributeInput, _ int) string {
		return attr.AttributeID
	})

	attrs, err := h.attrRepo.FindByIDsOrFail(ctx, attrIDs)
	if err != nil {
		return nil, err
	}

	attrMap := lo.KeyBy(attrs, func(a *attribute.Attribute) string {
		return a.ID
	})

	return lo.Map(inputs, func(attr CategoryAttributeInput, _ int) CategoryAttribute {
		slug := ""
		if a, ok := attrMap[attr.AttributeID]; ok {
			slug = a.Slug
		}
		return CategoryAttribute{
			AttributeID: attr.AttributeID,
			Slug:        slug,
			Role:        AttributeRole(attr.Role),
			SortOrder:   attr.SortOrder,
			Filterable:  attr.Filterable,
			Searchable:  attr.Searchable,
		}
	}), nil
}

func (h *createCategoryHandler) createCategory(cmd CreateCategoryCommand, attrs []CategoryAttribute) (*Category, error) {
	if cmd.ID != nil {
		return NewCategoryWithID(cmd.ID.String(), cmd.Name, cmd.Enabled, attrs)
	}
	return NewCategory(cmd.Name, cmd.Enabled, attrs)
}

func (h *createCategoryHandler) persistAndPublish(
	ctx context.Context,
	c *Category,
	msg outbox.Message,
) (*Category, error) {
	type createResult struct {
		Category *Category
		Send     outbox.SendFunc
	}

	res, err := mongo.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*createResult, error) {
		if err := h.repo.Insert(txCtx, c); err != nil {
			return nil, fmt.Errorf("failed to insert category: %w", err)
		}

		send, err := h.outbox.Create(txCtx, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to create outbox: %w", err)
		}

		return &createResult{
			Category: c,
			Send:     send,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	h.log(ctx).Debug("category created", zap.String("id", res.Category.ID))

	_ = res.Send(ctx) //nolint:errcheck // best-effort send, errors already logged in outbox

	return res.Category, nil
}

func (h *createCategoryHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "create-category-handler"))
}
