package category

import (
	"context"
	"fmt"

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
	outbox       outbox.Outbox
	txManager    persistence.TxManager
	eventFactory event.CategoryEventFactory
}

func NewCreateCategoryHandler(
	repo category.Repository,
	outbox outbox.Outbox,
	txManager persistence.TxManager,
	eventFactory event.CategoryEventFactory,
) CreateCategoryCommandHandler {
	return &createCategoryHandler{
		repo:         repo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *createCategoryHandler) Handle(ctx context.Context, cmd CreateCategoryCommand) (*category.Category, error) {
	var c *category.Category
	var err error

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

	if cmd.ID != nil {
		c, err = category.NewCategoryWithID(cmd.ID.String(), cmd.Name, cmd.Enabled, attributes)
	} else {
		c, err = category.NewCategory(cmd.Name, cmd.Enabled, attributes)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	type createResult struct {
		Category *category.Category
		Send     outbox.SendFunc
	}

	result, err := h.txManager.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		if err := h.repo.Insert(txCtx, c); err != nil {
			return nil, fmt.Errorf("failed to insert category: %w", err)
		}

		msg, err := h.eventFactory.NewCategoryCreatedOutboxMessage(txCtx, c)
		if err != nil {
			return nil, fmt.Errorf("failed to create event message: %w", err)
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

	res, ok := result.(*createResult)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	h.log(ctx).Debug("category created", zap.String("id", res.Category.ID))

	_ = res.Send(ctx)

	return res.Category, nil
}

func (h *createCategoryHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "create-category-handler"))
}
