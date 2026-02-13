package category

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
	categorymocks "github.com/Sokol111/ecommerce-catalog-service/internal/domain/category/mocks"
	eventmocks "github.com/Sokol111/ecommerce-catalog-service/internal/event/mocks"
	"github.com/Sokol111/ecommerce-catalog-service/internal/testutil/mocks"
	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
)

// mockSendFunc is a no-op send function for tests
func mockSendFunc(_ context.Context) error {
	return nil
}

// testCtx creates a context with a no-op logger for testing
func testCtx() context.Context {
	return logger.With(context.Background(), zap.NewNop())
}

// setupCreateCategoryHandler creates handler with mocked dependencies
func setupCreateCategoryHandler(t *testing.T) (
	*categorymocks.MockRepository,
	*attributemocks.MockRepository,
	*mocks.MockOutbox,
	*mocks.MockTxManager,
	*eventmocks.MockCategoryEventFactory,
	CreateCategoryCommandHandler,
) {
	repo := categorymocks.NewMockRepository(t)
	attrRepo := attributemocks.NewMockRepository(t)
	outboxMock := mocks.NewMockOutbox(t)
	txManager := mocks.NewMockTxManager(t)
	eventFactory := eventmocks.NewMockCategoryEventFactory(t)

	handler := NewCreateCategoryHandler(repo, attrRepo, outboxMock, txManager, eventFactory)

	return repo, attrRepo, outboxMock, txManager, eventFactory, handler
}

func TestCreateCategoryHandler_Handle_Success(t *testing.T) {
	repo, attrRepo, outboxMock, txManager, eventFactory, handler := setupCreateCategoryHandler(t)

	ctx := testCtx()
	cmd := CreateCategoryCommand{
		Name:    "Electronics",
		Enabled: true,
		Attributes: []CategoryAttributeInput{
			{
				AttributeID: "attr-1",
				Role:        "variant",
				Required:    true,
				SortOrder:   1,
				Filterable:  true,
				Searchable:  true,
			},
		},
	}

	// Mock attribute lookup
	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{"attr-1"}).
		Return([]*attribute.Attribute{
			attribute.Reconstruct("attr-1", 1, "Color", "color", attribute.AttributeTypeSingle, nil, true, nil, time.Now(), time.Now()),
		}, nil)

	// Mock event factory
	eventFactory.EXPECT().
		NewCategoryUpdatedOutboxMessage(mock.Anything, mock.AnythingOfType("*category.Category")).
		Return(outbox.Message{})

	// Mock transaction
	txManager.EXPECT().
		WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) (interface {}, error)")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	// Mock repository insert
	repo.EXPECT().
		Insert(mock.Anything, mock.AnythingOfType("*category.Category")).
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
	assert.True(t, result.Enabled)
	assert.Len(t, result.Attributes, 1)
	assert.Equal(t, "color", result.Attributes[0].Slug) // Slug from attribute
}

func TestCreateCategoryHandler_Handle_WithCustomID(t *testing.T) {
	repo, attrRepo, outboxMock, txManager, eventFactory, handler := setupCreateCategoryHandler(t)

	ctx := testCtx()
	customID := uuid.New()
	cmd := CreateCategoryCommand{
		ID:         &customID,
		Name:       "Electronics",
		Enabled:    false,
		Attributes: nil, // No attributes
	}

	// Mock empty attributes lookup
	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{}).
		Return([]*attribute.Attribute{}, nil)

	eventFactory.EXPECT().NewCategoryUpdatedOutboxMessage(mock.Anything, mock.Anything).Return(outbox.Message{})
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

func TestCreateCategoryHandler_Handle_InvalidName(t *testing.T) {
	_, attrRepo, _, _, _, handler := setupCreateCategoryHandler(t)

	ctx := testCtx()
	cmd := CreateCategoryCommand{
		Name:    "", // Invalid - empty name
		Enabled: false,
	}

	// Mock empty attributes lookup
	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{}).
		Return([]*attribute.Attribute{}, nil)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create category")
	assert.Nil(t, result)
}

func TestCreateCategoryHandler_Handle_AttributeNotFound(t *testing.T) {
	_, attrRepo, _, _, _, handler := setupCreateCategoryHandler(t)

	ctx := testCtx()
	cmd := CreateCategoryCommand{
		Name:    "Electronics",
		Enabled: false,
		Attributes: []CategoryAttributeInput{
			{AttributeID: "non-existent-attr"},
		},
	}

	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{"non-existent-attr"}).
		Return(nil, errors.New("attribute not found"))

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestCreateCategoryHandler_Handle_InsertError(t *testing.T) {
	repo, attrRepo, _, txManager, eventFactory, handler := setupCreateCategoryHandler(t)

	ctx := testCtx()
	cmd := CreateCategoryCommand{
		Name:    "Electronics",
		Enabled: false,
	}

	// Mock empty attributes lookup
	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{}).
		Return([]*attribute.Attribute{}, nil)

	eventFactory.EXPECT().NewCategoryUpdatedOutboxMessage(mock.Anything, mock.Anything).Return(outbox.Message{})
	txManager.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})
	repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(errors.New("database error"))

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert category")
	assert.Nil(t, result)
}

func TestCreateCategoryHandler_Handle_OutboxError(t *testing.T) {
	repo, attrRepo, outboxMock, txManager, eventFactory, handler := setupCreateCategoryHandler(t)

	ctx := testCtx()
	cmd := CreateCategoryCommand{
		Name:    "Electronics",
		Enabled: false,
	}

	// Mock empty attributes lookup
	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{}).
		Return([]*attribute.Attribute{}, nil)

	eventFactory.EXPECT().NewCategoryUpdatedOutboxMessage(mock.Anything, mock.Anything).Return(outbox.Message{})
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
