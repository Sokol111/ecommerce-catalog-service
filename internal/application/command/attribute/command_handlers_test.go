package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	attributemocks "github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute/mocks"
	eventmocks "github.com/Sokol111/ecommerce-catalog-service/internal/event/mocks"
	"github.com/Sokol111/ecommerce-catalog-service/internal/testutil/mocks"
	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

func ptr[T any](v T) *T {
	return &v
}

// mockSendFunc is a no-op send function for tests
func mockSendFunc(_ context.Context) error {
	return nil
}

// testCtx creates a context with a no-op logger for testing
func testCtx() context.Context {
	return logger.With(context.Background(), zap.NewNop())
}

// setupCreateAttributeHandler creates handler with mocked dependencies
func setupCreateAttributeHandler(t *testing.T) (
	*attributemocks.MockRepository,
	*mocks.MockOutbox,
	*mocks.MockTxManager,
	*eventmocks.MockAttributeEventFactory,
	CreateAttributeCommandHandler,
) {
	repo := attributemocks.NewMockRepository(t)
	outboxMock := mocks.NewMockOutbox(t)
	txManager := mocks.NewMockTxManager(t)
	eventFactory := eventmocks.NewMockAttributeEventFactory(t)

	handler := NewCreateAttributeHandler(repo, outboxMock, txManager, eventFactory)

	return repo, outboxMock, txManager, eventFactory, handler
}

func TestCreateAttributeHandler_Handle_Success(t *testing.T) {
	repo, outboxMock, txManager, eventFactory, handler := setupCreateAttributeHandler(t)

	ctx := logger.With(context.Background(), zap.NewNop())
	cmd := CreateAttributeCommand{
		Name:    "Color",
		Slug:    "color",
		Type:    "single",
		Unit:    nil,
		Enabled: true,
		Options: []OptionInput{
			{Name: "Red", Slug: "red", ColorCode: ptr("#FF0000"), SortOrder: 1},
			{Name: "Blue", Slug: "blue", ColorCode: ptr("#0000FF"), SortOrder: 2},
		},
	}

	// Mock event factory
	eventFactory.EXPECT().
		NewAttributeUpdatedOutboxMessage(mock.Anything, mock.AnythingOfType("*attribute.Attribute")).
		Return(outbox.Message{})

	// Mock transaction
	txManager.EXPECT().
		WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) (interface {}, error)")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	// Mock repository insert
	repo.EXPECT().
		Insert(mock.Anything, mock.AnythingOfType("*attribute.Attribute")).
		Return(nil)

	// Mock outbox create
	outboxMock.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(mockSendFunc, nil)

	// Execute
	result, err := handler.Handle(ctx, cmd)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, cmd.Name, result.Name)
	assert.Equal(t, cmd.Slug, result.Slug)
	assert.Equal(t, attribute.AttributeType(cmd.Type), result.Type)
	assert.True(t, result.Enabled)
	assert.Len(t, result.Options, 2)
}

func TestCreateAttributeHandler_Handle_WithCustomID(t *testing.T) {
	repo, outboxMock, txManager, eventFactory, handler := setupCreateAttributeHandler(t)

	ctx := testCtx()
	customID := uuid.New()
	cmd := CreateAttributeCommand{
		ID:      &customID,
		Name:    "Size",
		Slug:    "size",
		Type:    "multiple",
		Enabled: false,
	}

	eventFactory.EXPECT().NewAttributeUpdatedOutboxMessage(mock.Anything, mock.Anything).Return(outbox.Message{})
	txManager.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})
	repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil)
	outboxMock.EXPECT().Create(mock.Anything, mock.Anything).Return(mockSendFunc, nil)

	result, err := handler.Handle(ctx, cmd)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, customID.String(), result.ID)
}

func TestCreateAttributeHandler_Handle_InvalidName(t *testing.T) {
	_, _, _, _, handler := setupCreateAttributeHandler(t)

	ctx := testCtx()
	cmd := CreateAttributeCommand{
		Name:    "", // Invalid - empty name
		Slug:    "test",
		Type:    "single",
		Enabled: false,
	}

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create attribute")
	assert.Nil(t, result)
}

func TestCreateAttributeHandler_Handle_InvalidSlug(t *testing.T) {
	_, _, _, _, handler := setupCreateAttributeHandler(t)

	ctx := testCtx()
	cmd := CreateAttributeCommand{
		Name:    "Test",
		Slug:    "Invalid-Slug", // Invalid - uppercase
		Type:    "single",
		Enabled: false,
	}

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create attribute")
	assert.Nil(t, result)
}

func TestCreateAttributeHandler_Handle_InvalidType(t *testing.T) {
	_, _, _, _, handler := setupCreateAttributeHandler(t)

	ctx := testCtx()
	cmd := CreateAttributeCommand{
		Name:    "Test",
		Slug:    "test",
		Type:    "invalid-type", // Invalid type
		Enabled: false,
	}

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create attribute")
	assert.Nil(t, result)
}

func TestCreateAttributeHandler_Handle_InsertError(t *testing.T) {
	repo, _, txManager, eventFactory, handler := setupCreateAttributeHandler(t)

	ctx := testCtx()
	cmd := CreateAttributeCommand{
		Name:    "Color",
		Slug:    "color",
		Type:    "single",
		Enabled: false,
	}

	eventFactory.EXPECT().NewAttributeUpdatedOutboxMessage(mock.Anything, mock.Anything).Return(outbox.Message{})
	txManager.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})
	repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(errors.New("database error"))

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert attribute")
	assert.Nil(t, result)
}

func TestCreateAttributeHandler_Handle_OutboxError(t *testing.T) {
	repo, outboxMock, txManager, eventFactory, handler := setupCreateAttributeHandler(t)

	ctx := testCtx()
	cmd := CreateAttributeCommand{
		Name:    "Color",
		Slug:    "color",
		Type:    "single",
		Enabled: false,
	}

	eventFactory.EXPECT().NewAttributeUpdatedOutboxMessage(mock.Anything, mock.Anything).Return(outbox.Message{})
	txManager.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})
	repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil)
	outboxMock.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, errors.New("outbox error"))

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create outbox")
	assert.Nil(t, result)
}

func TestCreateAttributeHandler_Handle_AllAttributeTypes(t *testing.T) {
	tests := []struct {
		name     string
		attrType string
	}{
		{"single type", "single"},
		{"multiple type", "multiple"},
		{"range type", "range"},
		{"boolean type", "boolean"},
		{"text type", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, outboxMock, txManager, eventFactory, handler := setupCreateAttributeHandler(t)

			ctx := testCtx()
			cmd := CreateAttributeCommand{
				Name:    "Test Attr",
				Slug:    "test-attr",
				Type:    tt.attrType,
				Enabled: false,
			}

			eventFactory.EXPECT().NewAttributeUpdatedOutboxMessage(mock.Anything, mock.Anything).Return(outbox.Message{})
			txManager.EXPECT().
				WithTransaction(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
					return fn(ctx)
				})
			repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil)
			outboxMock.EXPECT().Create(mock.Anything, mock.Anything).Return(mockSendFunc, nil)

			result, err := handler.Handle(ctx, cmd)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, attribute.AttributeType(tt.attrType), result.Type)
		})
	}
}

// Helper to create a test attribute for update tests
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
