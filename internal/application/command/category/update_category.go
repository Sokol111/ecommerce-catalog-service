package category

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	"github.com/Sokol111/ecommerce-catalog-service/internal/event"
	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
	"go.uber.org/zap"
)

// UpdateCategoryCommand represents the input for updating a category
type UpdateCategoryCommand struct {
	ID         string
	Version    int
	Name       string
	Enabled    bool
	Attributes []CategoryAttributeInput
}

// UpdateCategoryCommandHandler defines the interface for updating categories
type UpdateCategoryCommandHandler interface {
	Handle(ctx context.Context, cmd UpdateCategoryCommand) (*category.Category, error)
}

type updateCategoryHandler struct {
	repo         category.Repository
	attrRepo     attribute.Repository
	outbox       outbox.Outbox
	txManager    persistence.TxManager
	eventFactory event.CategoryEventFactory
}

func NewUpdateCategoryHandler(
	repo category.Repository,
	attrRepo attribute.Repository,
	outbox outbox.Outbox,
	txManager persistence.TxManager,
	eventFactory event.CategoryEventFactory,
) UpdateCategoryCommandHandler {
	return &updateCategoryHandler{
		repo:         repo,
		attrRepo:     attrRepo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *updateCategoryHandler) Handle(ctx context.Context, cmd UpdateCategoryCommand) (*category.Category, error) {
	c, err := h.findAndValidateCategory(ctx, cmd.ID, cmd.Version)
	if err != nil {
		return nil, err
	}

	categoryAttrs, err := h.buildCategoryAttributes(ctx, cmd.Attributes)
	if err != nil {
		return nil, err
	}

	if err := c.Update(cmd.Name, cmd.Enabled, categoryAttrs); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return h.persistAndPublish(ctx, c)
}

func (h *updateCategoryHandler) findAndValidateCategory(ctx context.Context, id string, version int) (*category.Category, error) {
	c, err := h.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, persistence.ErrEntityNotFound) {
			return nil, persistence.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	if c.Version != version {
		return nil, persistence.ErrOptimisticLocking
	}

	return c, nil
}

func (h *updateCategoryHandler) buildCategoryAttributes(ctx context.Context, inputs []CategoryAttributeInput) ([]category.CategoryAttribute, error) {
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

	return lo.Map(inputs, func(attr CategoryAttributeInput, _ int) category.CategoryAttribute {
		slug := ""
		if a, ok := attrMap[attr.AttributeID]; ok {
			slug = a.Slug
		}
		return category.CategoryAttribute{
			AttributeID: attr.AttributeID,
			Slug:        slug,
			Role:        category.AttributeRole(attr.Role),
			Required:    attr.Required,
			SortOrder:   attr.SortOrder,
			Filterable:  attr.Filterable,
			Searchable:  attr.Searchable,
		}
	}), nil
}

func (h *updateCategoryHandler) persistAndPublish(
	ctx context.Context,
	c *category.Category,
) (*category.Category, error) {
	type updateResult struct {
		Category *category.Category
		Send     outbox.SendFunc
	}

	res, err := persistence.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*updateResult, error) {
		updated, err := h.repo.Update(txCtx, c)
		if err != nil {
			if errors.Is(err, persistence.ErrOptimisticLocking) {
				return nil, persistence.ErrOptimisticLocking
			}
			return nil, fmt.Errorf("failed to update category: %w", err)
		}

		msg := h.eventFactory.NewCategoryUpdatedOutboxMessage(txCtx, updated)

		send, err := h.outbox.Create(txCtx, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to create outbox: %w", err)
		}

		return &updateResult{
			Category: updated,
			Send:     send,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	h.log(ctx).Debug("category updated", zap.String("id", res.Category.ID))

	_ = res.Send(ctx)

	return res.Category, nil
}

func (h *updateCategoryHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "update-category-handler"))
}
