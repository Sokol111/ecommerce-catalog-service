package attribute

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
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
	Handle(ctx context.Context, cmd CreateAttributeCommand) (*Attribute, error)
}

type createAttributeHandler struct {
	repo         Repository
	outbox       outbox.Outbox
	txManager    mongo.TxManager
	eventFactory AttributeEventFactory
}

func NewCreateAttributeHandler(
	repo Repository,
	outbox outbox.Outbox,
	txManager mongo.TxManager,
	eventFactory AttributeEventFactory,
) CreateAttributeCommandHandler {
	return &createAttributeHandler{
		repo:         repo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *createAttributeHandler) Handle(ctx context.Context, cmd CreateAttributeCommand) (*Attribute, error) {
	options := lo.Map(cmd.Options, func(opt OptionInput, _ int) Option {
		return Option(opt)
	})

	var id string
	if cmd.ID != nil {
		id = cmd.ID.String()
	}

	a, err := NewAttribute(
		id,
		cmd.Name,
		cmd.Slug,
		AttributeType(cmd.Type),
		cmd.Unit,
		cmd.Enabled,
		options,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create attribute: %w", err)
	}

	msg := h.eventFactory.NewAttributeUpdatedOutboxMessage(ctx, a)

	return h.persistAndPublish(ctx, a, msg)
}

func (h *createAttributeHandler) persistAndPublish(
	ctx context.Context,
	a *Attribute,
	msg outbox.Message,
) (*Attribute, error) {
	type createResult struct {
		Attribute *Attribute
		Send      outbox.SendFunc
	}

	res, err := mongo.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*createResult, error) {
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

	_ = res.Send(ctx) //nolint:errcheck // best-effort send, errors already logged in outbox

	return res.Attribute, nil
}

func (h *createAttributeHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "create-attribute-handler"))
}
