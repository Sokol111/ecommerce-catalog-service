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
	// Collect attribute IDs for validation and enrichment
	attrIDs := lo.Map(cmd.Attributes, func(attr CategoryAttributeInput, _ int) string {
		return attr.AttributeID
	})

	// Fetch and validate attributes
	attrs, err := h.fetchAndValidateAttributes(ctx, attrIDs)
	if err != nil {
		return nil, err
	}

	// Convert attributes from command to domain
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

	// Create category
	c, err := h.createCategory(cmd, categoryAttrs)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	// Persist and publish
	return h.persistAndPublish(ctx, c, attrs)
}

func (h *createCategoryHandler) fetchAndValidateAttributes(ctx context.Context, attrIDs []string) ([]*attribute.Attribute, error) {
	if len(attrIDs) == 0 {
		return []*attribute.Attribute{}, nil
	}

	attrs, err := h.attrRepo.FindByIDs(ctx, attrIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attributes: %w", err)
	}

	// Validate all requested attributes exist
	foundIDs := lo.SliceToMap(attrs, func(attr *attribute.Attribute) (string, struct{}) {
		return attr.ID, struct{}{}
	})

	missingID, found := lo.Find(attrIDs, func(id string) bool {
		_, exists := foundIDs[id]
		return !exists
	})
	if found {
		return nil, fmt.Errorf("attribute not found: %s", missingID)
	}

	return attrs, nil
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
	attrs []*attribute.Attribute,
) (*category.Category, error) {
	type createResult struct {
		Category *category.Category
		Send     outbox.SendFunc
	}

	result, err := h.txManager.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		if err := h.repo.Insert(txCtx, c); err != nil {
			return nil, fmt.Errorf("failed to insert category: %w", err)
		}

		msg, err := h.eventFactory.NewCategoryCreatedOutboxMessage(txCtx, c, attrs)
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
