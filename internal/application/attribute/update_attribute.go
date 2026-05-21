package attribute

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"go.uber.org/zap"
)

type UpdateAttributeCommand struct {
	ID      string
	Version int
	Name    string
	Unit    *string
	Enabled bool
	Options []OptionInput
}

type UpdateAttributeCommandHandler interface {
	Handle(ctx context.Context, cmd UpdateAttributeCommand) (*Attribute, error)
}

type updateAttributeHandler struct {
	repo         Repository
	outbox       outbox.Outbox
	txManager    mongo.TxManager
	eventFactory AttributeEventFactory
}

func NewUpdateAttributeHandler(
	repo Repository,
	outbox outbox.Outbox,
	txManager mongo.TxManager,
	eventFactory AttributeEventFactory,
) UpdateAttributeCommandHandler {
	return &updateAttributeHandler{
		repo:         repo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *updateAttributeHandler) Handle(ctx context.Context, cmd UpdateAttributeCommand) (*Attribute, error) {
	a, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		if errors.Is(err, mongo.ErrEntityNotFound) {
			return nil, mongo.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}

	if a.Version != cmd.Version {
		return nil, mongo.ErrOptimisticLocking
	}

	options := lo.Map(cmd.Options, func(opt OptionInput, _ int) Option {
		return Option{
			Name:      opt.Name,
			Slug:      opt.Slug,
			ColorCode: opt.ColorCode,
			SortOrder: opt.SortOrder,
		}
	})

	if err := a.Update(
		cmd.Name,
		cmd.Unit,
		cmd.Enabled,
		options,
	); err != nil {
		return nil, fmt.Errorf("failed to update attribute: %w", err)
	}

	return h.persistAndPublish(ctx, a)
}

func (h *updateAttributeHandler) persistAndPublish(
	ctx context.Context,
	a *Attribute,
) (*Attribute, error) {
	type updateResult struct {
		Attribute *Attribute
		Send      outbox.SendFunc
	}

	res, err := mongo.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*updateResult, error) {
		updated, err := h.repo.Update(txCtx, a)
		if err != nil {
			if errors.Is(err, mongo.ErrOptimisticLocking) {
				return nil, mongo.ErrOptimisticLocking
			}
			return nil, fmt.Errorf("failed to update attribute: %w", err)
		}

		msg := h.eventFactory.NewAttributeUpdatedOutboxMessage(txCtx, updated)

		send, err := h.outbox.Create(txCtx, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to create outbox: %w", err)
		}

		return &updateResult{
			Attribute: updated,
			Send:      send,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	h.log(ctx).Debug("attribute updated", zap.String("id", res.Attribute.ID))

	_ = res.Send(ctx) //nolint:errcheck // best-effort send, errors already logged in outbox

	return res.Attribute, nil
}

func (h *updateAttributeHandler) log(ctx context.Context) *zap.Logger {
	return logger.Get(ctx).With(zap.String("component", "update-attribute-handler"))
}
