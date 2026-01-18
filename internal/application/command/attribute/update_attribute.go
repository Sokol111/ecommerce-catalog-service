package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
)

type UpdateAttributeCommand struct {
	ID      string
	Version int
	Name    string
	Slug    string
	Type    string
	Unit    *string
	Enabled bool
	Options []OptionInput
}

type UpdateAttributeCommandHandler interface {
	Handle(ctx context.Context, cmd UpdateAttributeCommand) (*attribute.Attribute, error)
}

type updateAttributeHandler struct {
	repo attribute.Repository
}

func NewUpdateAttributeHandler(repo attribute.Repository) UpdateAttributeCommandHandler {
	return &updateAttributeHandler{
		repo: repo,
	}
}

func (h *updateAttributeHandler) Handle(ctx context.Context, cmd UpdateAttributeCommand) (*attribute.Attribute, error) {
	a, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		if errors.Is(err, persistence.ErrEntityNotFound) {
			return nil, persistence.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}

	if a.Version != cmd.Version {
		return nil, persistence.ErrOptimisticLocking
	}

	attrType := attribute.AttributeType(cmd.Type)

	options := lo.Map(cmd.Options, func(opt OptionInput, _ int) attribute.Option {
		return attribute.Option{
			Name:      opt.Name,
			Slug:      opt.Slug,
			ColorCode: opt.ColorCode,
			SortOrder: opt.SortOrder,
			Enabled:   opt.Enabled,
		}
	})

	if err := a.Update(
		cmd.Name,
		cmd.Slug,
		attrType,
		cmd.Unit,
		cmd.Enabled,
		options,
	); err != nil {
		return nil, fmt.Errorf("failed to update attribute: %w", err)
	}

	updated, err := h.repo.Update(ctx, a)
	if err != nil {
		if !errors.Is(err, persistence.ErrOptimisticLocking) {
			return nil, fmt.Errorf("failed to update attribute: %w", err)
		}
		return nil, err
	}

	return updated, nil
}
