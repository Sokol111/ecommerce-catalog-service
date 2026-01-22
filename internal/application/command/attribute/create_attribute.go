package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/event"
	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
	"go.uber.org/zap"
)

type OptionInput struct {
	Name      string
	Slug      string
	ColorCode *string
	SortOrder int
}

type CreateAttributeCommand struct {
	ID      *uuid.UUID
	Name    string
	Slug    string
	Type    string
	Unit    *string
	Enabled bool
	Options []OptionInput
}

type CreateAttributeCommandHandler interface {
	Handle(ctx context.Context, cmd CreateAttributeCommand) (*attribute.Attribute, error)
}

type createAttributeHandler struct {
	repo         attribute.Repository
	outbox       outbox.Outbox
	txManager    persistence.TxManager
	eventFactory event.AttributeEventFactory
}

func NewCreateAttributeHandler(
	repo attribute.Repository,
	outbox outbox.Outbox,
	txManager persistence.TxManager,
	eventFactory event.AttributeEventFactory,
) CreateAttributeCommandHandler {
	return &createAttributeHandler{
		repo:         repo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *createAttributeHandler) Handle(ctx context.Context, cmd CreateAttributeCommand) (*attribute.Attribute, error) {
	options := lo.Map(cmd.Options, func(opt OptionInput, _ int) attribute.Option {
		return attribute.Option{
			Name:      opt.Name,
			Slug:      opt.Slug,
			ColorCode: opt.ColorCode,
			SortOrder: opt.SortOrder,
		}
	})

	var id string
	if cmd.ID != nil {
		id = cmd.ID.String()
	}

	a, err := attribute.NewAttribute(
		id,
		cmd.Name,
		cmd.Slug,
		attribute.AttributeType(cmd.Type),
		cmd.Unit,
		cmd.Enabled,
		options,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create attribute: %w", err)
	}

	msg, err := h.eventFactory.NewAttributeUpdatedOutboxMessage(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("failed to create event message: %w", err)
	}

	return h.persistAndPublish(ctx, a, msg)
}

func (h *createAttributeHandler) persistAndPublish(
	ctx context.Context,
	a *attribute.Attribute,
	msg outbox.Message,
) (*attribute.Attribute, error) {
	type createResult struct {
		Attribute *attribute.Attribute
		Send      outbox.SendFunc
	}

	res, err := persistence.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*createResult, error) {
		if err := h.repo.Insert(txCtx, a); err != nil {
			return nil, fmt.Errorf("failed to insert attribute: %w", err)
		}

		send, err := h.outbox.Create(txCtx, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to create outbox: %w", err)
		}

		return &createResult{
			Attribute: a,
			Send:      send,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	h.log(ctx).Debug("attribute created", zap.String("id", res.Attribute.ID))

	_ = res.Send(ctx)

	return res.Attribute, nil
}

func (h *createAttributeHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "create-attribute-handler"))
}
