package product

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

	attributemocks "github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute/mocks"
	categorymocks "github.com/Sokol111/ecommerce-catalog-service/internal/domain/category/mocks"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	productmocks "github.com/Sokol111/ecommerce-catalog-service/internal/domain/product/mocks"
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

// setupCreateProductHandler creates handler with mocked dependencies
func setupCreateProductHandler(t *testing.T) (
	*productmocks.MockRepository,
	*attributemocks.MockRepository,
	*categorymocks.MockRepository,
	*mocks.MockOutbox,
	*mocks.MockTxManager,
	*eventmocks.MockProductEventFactory,
	CreateProductCommandHandler,
) {
	repo := productmocks.NewMockRepository(t)
	attrRepo := attributemocks.NewMockRepository(t)
	categoryRepo := categorymocks.NewMockRepository(t)
	outboxMock := mocks.NewMockOutbox(t)
	txManager := mocks.NewMockTxManager(t)
	eventFactory := eventmocks.NewMockProductEventFactory(t)

	handler := NewCreateProductHandler(repo, attrRepo, categoryRepo, outboxMock, txManager, eventFactory)

	return repo, attrRepo, categoryRepo, outboxMock, txManager, eventFactory, handler
}

func TestCreateProductHandler_Handle_Success(t *testing.T) {
	repo, attrRepo, categoryRepo, outboxMock, txManager, eventFactory, handler := setupCreateProductHandler(t)

	ctx := testCtx()
	categoryID := "category-123"
	cmd := CreateProductCommand{
		Name:        "Test Product",
		Description: ptr("Test description"),
		Price:       99.99,
		Quantity:    10,
		ImageID:     ptr("image-123"),
		CategoryID:  &categoryID,
		Enabled:     true,
		Attributes:  nil,
	}

	// Mock category exists check
	categoryRepo.EXPECT().
		Exists(mock.Anything, categoryID).
		Return(true, nil)

	// Mock empty attributes lookup
	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{}).
		Return(nil, nil)

	// Mock event factory
	eventFactory.EXPECT().
		NewProductUpdatedOutboxMessage(mock.Anything, mock.AnythingOfType("*product.Product")).
		Return(outbox.Message{})

	// Mock transaction - execute the function passed to it
	txManager.EXPECT().
		WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) (interface {}, error)")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	// Mock repository insert
	repo.EXPECT().
		Insert(mock.Anything, mock.AnythingOfType("*product.Product")).
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
	assert.Equal(t, cmd.Description, result.Description)
	assert.Equal(t, cmd.Price, result.Price)
	assert.Equal(t, cmd.Quantity, result.Quantity)
	assert.Equal(t, cmd.CategoryID, result.CategoryID)
	assert.True(t, result.Enabled)
}

func TestCreateProductHandler_Handle_WithCustomID(t *testing.T) {
	repo, attrRepo, categoryRepo, outboxMock, txManager, eventFactory, handler := setupCreateProductHandler(t)

	ctx := testCtx()
	customID := uuid.New()
	categoryID := "category-123"
	cmd := CreateProductCommand{
		ID:         &customID,
		Name:       "Test Product",
		Price:      99.99,
		Quantity:   10,
		ImageID:    ptr("image-123"),
		CategoryID: &categoryID,
		Enabled:    true,
	}

	categoryRepo.EXPECT().Exists(mock.Anything, categoryID).Return(true, nil)
	attrRepo.EXPECT().FindByIDsOrFail(mock.Anything, []string{}).Return(nil, nil)
	eventFactory.EXPECT().NewProductUpdatedOutboxMessage(mock.Anything, mock.Anything).Return(outbox.Message{})
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

func TestCreateProductHandler_Handle_CategoryNotFound(t *testing.T) {
	_, _, categoryRepo, _, _, _, handler := setupCreateProductHandler(t)

	ctx := testCtx()
	categoryID := "non-existent-category"
	cmd := CreateProductCommand{
		Name:       "Test Product",
		Price:      10,
		Quantity:   5,
		CategoryID: &categoryID,
		Enabled:    false,
	}

	categoryRepo.EXPECT().
		Exists(mock.Anything, categoryID).
		Return(false, nil)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, product.ErrCategoryNotFound)
	assert.Nil(t, result)
}

func TestCreateProductHandler_Handle_CategoryCheckError(t *testing.T) {
	_, _, categoryRepo, _, _, _, handler := setupCreateProductHandler(t)

	ctx := testCtx()
	categoryID := "category-123"
	cmd := CreateProductCommand{
		Name:       "Test Product",
		Price:      10,
		Quantity:   5,
		CategoryID: &categoryID,
		Enabled:    false,
	}

	categoryRepo.EXPECT().
		Exists(mock.Anything, categoryID).
		Return(false, errors.New("database error"))

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check category")
	assert.Nil(t, result)
}

func TestCreateProductHandler_Handle_InvalidProductData(t *testing.T) {
	_, attrRepo, _, _, _, _, handler := setupCreateProductHandler(t)

	ctx := testCtx()
	cmd := CreateProductCommand{
		Name:     "", // Invalid - empty name
		Price:    10,
		Quantity: 5,
		Enabled:  false,
	}

	// Mock empty attributes lookup
	attrRepo.EXPECT().FindByIDsOrFail(mock.Anything, []string{}).Return(nil, nil)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create product")
	assert.Nil(t, result)
}

func TestCreateProductHandler_Handle_AttributeValidationFailure(t *testing.T) {
	_, attrRepo, _, _, _, _, handler := setupCreateProductHandler(t)

	ctx := testCtx()
	cmd := CreateProductCommand{
		Name:     "Test Product",
		Price:    10,
		Quantity: 5,
		Enabled:  false,
		Attributes: []product.AttributeValue{
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

func TestCreateProductHandler_Handle_InsertError(t *testing.T) {
	repo, attrRepo, categoryRepo, outboxMock, txManager, eventFactory, handler := setupCreateProductHandler(t)

	ctx := testCtx()
	categoryID := "category-123"
	cmd := CreateProductCommand{
		Name:       "Test Product",
		Price:      99.99,
		Quantity:   10,
		ImageID:    ptr("image-123"),
		CategoryID: &categoryID,
		Enabled:    true,
	}

	categoryRepo.EXPECT().Exists(mock.Anything, categoryID).Return(true, nil)
	attrRepo.EXPECT().FindByIDsOrFail(mock.Anything, []string{}).Return(nil, nil)
	eventFactory.EXPECT().NewProductUpdatedOutboxMessage(mock.Anything, mock.Anything).Return(outbox.Message{})
	txManager.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})
	repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(errors.New("database error"))

	// Outbox mock should not be called since Insert fails
	_ = outboxMock

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert product")
	assert.Nil(t, result)
}

func TestCreateProductHandler_Handle_OutboxError(t *testing.T) {
	repo, attrRepo, categoryRepo, outboxMock, txManager, eventFactory, handler := setupCreateProductHandler(t)

	ctx := testCtx()
	categoryID := "category-123"
	cmd := CreateProductCommand{
		Name:       "Test Product",
		Price:      99.99,
		Quantity:   10,
		ImageID:    ptr("image-123"),
		CategoryID: &categoryID,
		Enabled:    true,
	}

	categoryRepo.EXPECT().Exists(mock.Anything, categoryID).Return(true, nil)
	attrRepo.EXPECT().FindByIDsOrFail(mock.Anything, []string{}).Return(nil, nil)
	eventFactory.EXPECT().NewProductUpdatedOutboxMessage(mock.Anything, mock.Anything).Return(outbox.Message{})
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

func TestCreateProductHandler_Handle_NoCategoryValidation(t *testing.T) {
	repo, attrRepo, _, outboxMock, txManager, eventFactory, handler := setupCreateProductHandler(t)

	ctx := testCtx()
	cmd := CreateProductCommand{
		Name:       "Test Product",
		Price:      0,
		Quantity:   0,
		CategoryID: nil, // No category - should skip validation
		Enabled:    false,
	}

	// Category repo should NOT be called
	attrRepo.EXPECT().FindByIDsOrFail(mock.Anything, []string{}).Return(nil, nil)
	eventFactory.EXPECT().NewProductUpdatedOutboxMessage(mock.Anything, mock.Anything).Return(outbox.Message{})
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
	assert.Nil(t, result.CategoryID)
}

// Test helper to create a product for update tests
func createTestProduct() *product.Product {
	return product.Reconstruct(
		"product-123",
		1,
		"Original Product",
		ptr("Original description"),
		99.99,
		10,
		ptr("image-123"),
		ptr("category-123"),
		true,
		nil,
		time.Now().UTC(),
		time.Now().UTC(),
	)
}
