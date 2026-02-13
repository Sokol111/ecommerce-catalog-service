package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	attributemocks "github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute/mocks"
	eventmocks "github.com/Sokol111/ecommerce-catalog-service/internal/event/mocks"
	"github.com/Sokol111/ecommerce-catalog-service/internal/testutil/mocks"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

// createTestAttribute creates a test attribute for update tests
func createTestAttribute() *attribute.Attribute {
	return attribute.Reconstruct(
		"attr-123",
		1,
		"Original Name",
		"original-slug",
		attribute.AttributeTypeSingle,
		nil,
		false,
		[]attribute.Option{
			{Name: "Option 1", Slug: "option-1", SortOrder: 1},
		},
		time.Now().UTC(),
		time.Now().UTC(),
	)
}

// setupUpdateAttributeHandler creates handler with mocked dependencies
func setupUpdateAttributeHandler(t *testing.T) (
	*attributemocks.MockRepository,
	*mocks.MockOutbox,
	*mocks.MockTxManager,
	*eventmocks.MockAttributeEventFactory,
	UpdateAttributeCommandHandler,
) {
	repo := attributemocks.NewMockRepository(t)
	outboxMock := mocks.NewMockOutbox(t)
	txManager := mocks.NewMockTxManager(t)
	eventFactory := eventmocks.NewMockAttributeEventFactory(t)

	handler := NewUpdateAttributeHandler(repo, outboxMock, txManager, eventFactory)

	return repo, outboxMock, txManager, eventFactory, handler
}

func TestUpdateAttributeHandler_Handle_Success(t *testing.T) {
	repo, outboxMock, txManager, eventFactory, handler := setupUpdateAttributeHandler(t)

	ctx := testCtx()
	existingAttr := createTestAttribute()

	cmd := UpdateAttributeCommand{
		ID:      existingAttr.ID,
		Version: existingAttr.Version,
		Name:    "Updated Name",
		Unit:    ptr("cm"),
		Enabled: true,
		Options: []OptionInput{
			{Name: "New Option", Slug: "new-option", SortOrder: 1},
		},
	}

	// Mock find existing attribute
	repo.EXPECT().
		FindByID(mock.Anything, existingAttr.ID).
		Return(existingAttr, nil)

	// Mock transaction
	txManager.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	// Mock update
	repo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*attribute.Attribute")).
		RunAndReturn(func(_ context.Context, a *attribute.Attribute) (*attribute.Attribute, error) {
			return a, nil
		})

	// Mock event factory
	eventFactory.EXPECT().
		NewAttributeUpdatedOutboxMessage(mock.Anything, mock.Anything).
		Return(outbox.Message{})

	// Mock outbox
	outboxMock.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(mockSendFunc, nil)

	// Execute
	result, err := handler.Handle(ctx, cmd)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, cmd.Name, result.Name)
	assert.Equal(t, cmd.Unit, result.Unit)
	assert.True(t, result.Enabled)
	// Slug and Type should remain unchanged
	assert.Equal(t, existingAttr.Slug, result.Slug)
	assert.Equal(t, existingAttr.Type, result.Type)
}

func TestUpdateAttributeHandler_Handle_NotFound(t *testing.T) {
	repo, _, _, _, handler := setupUpdateAttributeHandler(t)

	ctx := testCtx()
	cmd := UpdateAttributeCommand{
		ID:      "non-existent-id",
		Version: 1,
		Name:    "Updated Name",
	}

	repo.EXPECT().
		FindByID(mock.Anything, cmd.ID).
		Return(nil, mongo.ErrEntityNotFound)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, mongo.ErrEntityNotFound)
	assert.Nil(t, result)
}

func TestUpdateAttributeHandler_Handle_OptimisticLockingVersionMismatch(t *testing.T) {
	repo, _, _, _, handler := setupUpdateAttributeHandler(t)

	ctx := testCtx()
	existingAttr := createTestAttribute() // Version 1

	cmd := UpdateAttributeCommand{
		ID:      existingAttr.ID,
		Version: 2, // Wrong version
		Name:    "Updated Name",
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingAttr.ID).
		Return(existingAttr, nil)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, mongo.ErrOptimisticLocking)
	assert.Nil(t, result)
}

func TestUpdateAttributeHandler_Handle_InvalidName(t *testing.T) {
	repo, _, _, _, handler := setupUpdateAttributeHandler(t)

	ctx := testCtx()
	existingAttr := createTestAttribute()

	cmd := UpdateAttributeCommand{
		ID:      existingAttr.ID,
		Version: existingAttr.Version,
		Name:    "", // Invalid - empty name
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingAttr.ID).
		Return(existingAttr, nil)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update attribute")
	assert.Nil(t, result)
}

func TestUpdateAttributeHandler_Handle_InvalidOptions(t *testing.T) {
	repo, _, _, _, handler := setupUpdateAttributeHandler(t)

	ctx := testCtx()
	existingAttr := createTestAttribute()

	cmd := UpdateAttributeCommand{
		ID:      existingAttr.ID,
		Version: existingAttr.Version,
		Name:    "Valid Name",
		Options: []OptionInput{
			{Name: "", Slug: "option", SortOrder: 1}, // Invalid - empty name
		},
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingAttr.ID).
		Return(existingAttr, nil)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update attribute")
	assert.Nil(t, result)
}

func TestUpdateAttributeHandler_Handle_UpdateRepositoryError(t *testing.T) {
	repo, _, txManager, _, handler := setupUpdateAttributeHandler(t)

	ctx := testCtx()
	existingAttr := createTestAttribute()

	cmd := UpdateAttributeCommand{
		ID:      existingAttr.ID,
		Version: existingAttr.Version,
		Name:    "Updated Name",
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingAttr.ID).
		Return(existingAttr, nil)

	txManager.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	repo.EXPECT().
		Update(mock.Anything, mock.Anything).
		Return(nil, errors.New("database error"))

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update attribute")
	assert.Nil(t, result)
}
