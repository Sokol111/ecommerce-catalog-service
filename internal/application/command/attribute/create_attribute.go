package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
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
	repo attribute.Repository
}

func NewCreateAttributeHandler(repo attribute.Repository) CreateAttributeCommandHandler {
	return &createAttributeHandler{
		repo: repo,
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

	if err := h.repo.Insert(ctx, a); err != nil {
		return nil, fmt.Errorf("failed to insert attribute: %w", err)
	}

	return a, nil
}
