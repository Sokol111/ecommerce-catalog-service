package category

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	"github.com/Sokol111/ecommerce-catalog-service/internal/event"
	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CategoryAttributeInput represents the input for a category attribute
type CategoryAttributeInput struct {
	AttributeID string
	Role        string
	Required    bool
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
	Handle(ctx context.Context, cmd CreateCategoryCommand) (*category.Category, error)
}

type createCategoryHandler struct {
	repo         category.Repository
	attrRepo     attribute.Repository
	outbox       outbox.Outbox
	txManager    persistence.TxManager
	eventFactory event.CategoryEventFactory
}

func NewCreateCategoryHandler(
	repo category.Repository,
	attrRepo attribute.Repository,
	outbox outbox.Outbox,
	txManager persistence.TxManager,
	eventFactory event.CategoryEventFactory,
) CreateCategoryCommandHandler {
	return &createCategoryHandler{
		repo:         repo,
		attrRepo:     attrRepo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *createCategoryHandler) Handle(ctx context.Context, cmd CreateCategoryCommand) (*category.Category, error) {
	attrIDs := lo.Map(cmd.Attributes, func(attr CategoryAttributeInput, _ int) string {
		return attr.AttributeID
	})

	attrs, err := h.attrRepo.FindByIDsOrFail(ctx, attrIDs)
	if err != nil {
		return nil, err
	}

	categoryAttrs := lo.Map(cmd.Attributes, func(attr CategoryAttributeInput, _ int) category.CategoryAttribute {
		return category.CategoryAttribute{
			AttributeID: attr.AttributeID,
			Role:        category.AttributeRole(attr.Role),
			Required:    attr.Required,
			SortOrder:   attr.SortOrder,
			Filterable:  attr.Filterable,
			Searchable:  attr.Searchable,
		}
	})

	c, err := h.createCategory(cmd, categoryAttrs)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	msg, err := h.eventFactory.NewCategoryCreatedOutboxMessage(ctx, c, attrs)
	if err != nil {
		return nil, fmt.Errorf("failed to create event message: %w", err)
	}

	return h.persistAndPublish(ctx, c, msg)
}

func (h *createCategoryHandler) createCategory(cmd CreateCategoryCommand, attrs []category.CategoryAttribute) (*category.Category, error) {
	if cmd.ID != nil {
		return category.NewCategoryWithID(cmd.ID.String(), cmd.Name, cmd.Enabled, attrs)
	}
	return category.NewCategory(cmd.Name, cmd.Enabled, attrs)
}

func (h *createCategoryHandler) persistAndPublish(
	ctx context.Context,
	c *category.Category,
	msg outbox.Message,
) (*category.Category, error) {
	type createResult struct {
		Category *category.Category
		Send     outbox.SendFunc
	}

	res, err := persistence.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*createResult, error) {
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

	_ = res.Send(ctx)

	return res.Category, nil
}

func (h *createCategoryHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "create-category-handler"))
}
