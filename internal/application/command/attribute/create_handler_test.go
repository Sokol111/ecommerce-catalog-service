package command

import (
	"context"
	"errors"
	"testing"

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
