package category

import (
	"context"
	"errors"
	"fmt"

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
	outbox       outbox.Outbox
	txManager    persistence.TxManager
	eventFactory event.CategoryEventFactory
}

func NewUpdateCategoryHandler(
	repo category.Repository,
	outbox outbox.Outbox,
	txManager persistence.TxManager,
	eventFactory event.CategoryEventFactory,
) UpdateCategoryCommandHandler {
	return &updateCategoryHandler{
		repo:         repo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *updateCategoryHandler) Handle(ctx context.Context, cmd UpdateCategoryCommand) (*category.Category, error) {
	c, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		if errors.Is(err, persistence.ErrEntityNotFound) {
			return nil, persistence.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	if c.Version != cmd.Version {
		return nil, persistence.ErrOptimisticLocking
	}

	// Convert attributes from command to domain
	attributes := make([]category.CategoryAttribute, 0, len(cmd.Attributes))
	for _, attr := range cmd.Attributes {
		attributes = append(attributes, category.CategoryAttribute{
			AttributeID: attr.AttributeID,
			Role:        category.AttributeRole(attr.Role),
			Required:    attr.Required,
			SortOrder:   attr.SortOrder,
			Filterable:  attr.Filterable,
			Searchable:  attr.Searchable,
		})
	}

	if err := c.Update(cmd.Name, cmd.Enabled, attributes); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	type updateResult struct {
		Category *category.Category
		Send     outbox.SendFunc
	}

	result, err := h.txManager.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		// Update in repository (with optimistic locking)
		updated, err := h.repo.Update(txCtx, c)
		if err != nil {
			if errors.Is(err, persistence.ErrOptimisticLocking) {
				return nil, persistence.ErrOptimisticLocking
			}
			return nil, fmt.Errorf("failed to update category: %w", err)
		}

		msg, err := h.eventFactory.NewCategoryUpdatedOutboxMessage(txCtx, updated)
		if err != nil {
			return nil, fmt.Errorf("failed to create event message: %w", err)
		}

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

	res, ok := result.(*updateResult)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	h.log(ctx).Debug("category updated", zap.String("id", res.Category.ID))

	_ = res.Send(ctx)

	return res.Category, nil
}

func (h *updateCategoryHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "update-category-handler"))
}
