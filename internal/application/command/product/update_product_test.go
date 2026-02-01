package product

import (
	"context"
	"errors"
	"testing"

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
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
)

// testCtx creates a context with a no-op logger for testing
func testCtxUpdate() context.Context {
	return logger.With(context.Background(), zap.NewNop())
}

// setupUpdateProductHandler creates handler with mocked dependencies
func setupUpdateProductHandler(t *testing.T) (
	*productmocks.MockRepository,
	*attributemocks.MockRepository,
	*categorymocks.MockRepository,
	*mocks.MockOutbox,
	*mocks.MockTxManager,
	*eventmocks.MockProductEventFactory,
	UpdateProductCommandHandler,
) {
	repo := productmocks.NewMockRepository(t)
	attrRepo := attributemocks.NewMockRepository(t)
	categoryRepo := categorymocks.NewMockRepository(t)
	outboxMock := mocks.NewMockOutbox(t)
	txManager := mocks.NewMockTxManager(t)
	eventFactory := eventmocks.NewMockProductEventFactory(t)

	handler := NewUpdateProductHandler(repo, attrRepo, categoryRepo, outboxMock, txManager, eventFactory)

	return repo, attrRepo, categoryRepo, outboxMock, txManager, eventFactory, handler
}

func TestUpdateProductHandler_Handle_Success(t *testing.T) {
	repo, attrRepo, categoryRepo, outboxMock, txManager, eventFactory, handler := setupUpdateProductHandler(t)

	ctx := testCtxUpdate()
	existingProduct := createTestProduct()
	categoryID := "category-456"

	cmd := UpdateProductCommand{
		ID:          existingProduct.ID,
		Version:     existingProduct.Version,
		Name:        "Updated Product",
		Description: ptr("Updated description"),
		Price:       199.99,
		Quantity:    20,
		ImageID:     ptr("image-456"),
		CategoryID:  &categoryID,
		Enabled:     true,
		Attributes:  nil,
	}

	// Mock find existing product
	repo.EXPECT().
		FindByID(mock.Anything, existingProduct.ID).
		Return(existingProduct, nil)

	// Mock category validation
	categoryRepo.EXPECT().
		Exists(mock.Anything, categoryID).
		Return(true, nil)

	// Mock empty attributes lookup
	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{}).
		Return(nil, nil)

	// Mock transaction
	txManager.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	// Mock update - return a copy of the updated product
	repo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*product.Product")).
		RunAndReturn(func(_ context.Context, p *product.Product) (*product.Product, error) {
			return p, nil
		})

	// Mock event factory
	eventFactory.EXPECT().
		NewProductUpdatedOutboxMessage(mock.Anything, mock.Anything).
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
	assert.Equal(t, cmd.Description, result.Description)
	assert.Equal(t, cmd.Price, result.Price)
	assert.Equal(t, cmd.Quantity, result.Quantity)
}

func TestUpdateProductHandler_Handle_ProductNotFound(t *testing.T) {
	repo, _, _, _, _, _, handler := setupUpdateProductHandler(t)

	ctx := testCtxUpdate()
	cmd := UpdateProductCommand{
		ID:      "non-existent-id",
		Version: 1,
		Name:    "Updated Product",
		Price:   99.99,
	}

	repo.EXPECT().
		FindByID(mock.Anything, cmd.ID).
		Return(nil, persistence.ErrEntityNotFound)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, persistence.ErrEntityNotFound)
	assert.Nil(t, result)
}

func TestUpdateProductHandler_Handle_OptimisticLockingVersionMismatch(t *testing.T) {
	repo, _, _, _, _, _, handler := setupUpdateProductHandler(t)

	ctx := testCtxUpdate()
	existingProduct := createTestProduct() // Version 1

	cmd := UpdateProductCommand{
		ID:      existingProduct.ID,
		Version: 2, // Wrong version
		Name:    "Updated Product",
		Price:   99.99,
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingProduct.ID).
		Return(existingProduct, nil)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, persistence.ErrOptimisticLocking)
	assert.Nil(t, result)
}

func TestUpdateProductHandler_Handle_CategoryNotFound(t *testing.T) {
	repo, _, categoryRepo, _, _, _, handler := setupUpdateProductHandler(t)

	ctx := testCtxUpdate()
	existingProduct := createTestProduct()
	categoryID := "non-existent-category"

	cmd := UpdateProductCommand{
		ID:         existingProduct.ID,
		Version:    existingProduct.Version,
		Name:       "Updated Product",
		Price:      99.99,
		Quantity:   10,
		ImageID:    ptr("image-123"),
		CategoryID: &categoryID,
		Enabled:    true,
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingProduct.ID).
		Return(existingProduct, nil)

	categoryRepo.EXPECT().
		Exists(mock.Anything, categoryID).
		Return(false, nil)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, product.ErrCategoryNotFound)
	assert.Nil(t, result)
}

func TestUpdateProductHandler_Handle_InvalidUpdateData(t *testing.T) {
	repo, attrRepo, categoryRepo, _, _, _, handler := setupUpdateProductHandler(t)

	ctx := testCtxUpdate()
	existingProduct := createTestProduct()
	categoryID := "category-123"

	cmd := UpdateProductCommand{
		ID:         existingProduct.ID,
		Version:    existingProduct.Version,
		Name:       "", // Invalid - empty name
		Price:      99.99,
		Quantity:   10,
		ImageID:    ptr("image-123"),
		CategoryID: &categoryID,
		Enabled:    true,
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingProduct.ID).
		Return(existingProduct, nil)

	categoryRepo.EXPECT().
		Exists(mock.Anything, categoryID).
		Return(true, nil)

	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{}).
		Return(nil, nil)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update product")
	assert.Nil(t, result)
}

func TestUpdateProductHandler_Handle_UpdateRepositoryError(t *testing.T) {
	repo, attrRepo, categoryRepo, _, txManager, _, handler := setupUpdateProductHandler(t)

	ctx := testCtxUpdate()
	existingProduct := createTestProduct()
	categoryID := "category-123"

	cmd := UpdateProductCommand{
		ID:         existingProduct.ID,
		Version:    existingProduct.Version,
		Name:       "Updated Product",
		Price:      199.99,
		Quantity:   20,
		ImageID:    ptr("image-123"),
		CategoryID: &categoryID,
		Enabled:    true,
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingProduct.ID).
		Return(existingProduct, nil)

	categoryRepo.EXPECT().
		Exists(mock.Anything, categoryID).
		Return(true, nil)

	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{}).
		Return(nil, nil)

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
	assert.Contains(t, err.Error(), "failed to update product")
	assert.Nil(t, result)
}

func TestUpdateProductHandler_Handle_OptimisticLockingOnUpdate(t *testing.T) {
	repo, attrRepo, categoryRepo, _, txManager, _, handler := setupUpdateProductHandler(t)

	ctx := testCtxUpdate()
	existingProduct := createTestProduct()
	categoryID := "category-123"

	cmd := UpdateProductCommand{
		ID:         existingProduct.ID,
		Version:    existingProduct.Version,
		Name:       "Updated Product",
		Price:      199.99,
		Quantity:   20,
		ImageID:    ptr("image-123"),
		CategoryID: &categoryID,
		Enabled:    true,
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingProduct.ID).
		Return(existingProduct, nil)

	categoryRepo.EXPECT().
		Exists(mock.Anything, categoryID).
		Return(true, nil)

	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{}).
		Return(nil, nil)

	txManager.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	repo.EXPECT().
		Update(mock.Anything, mock.Anything).
		Return(nil, persistence.ErrOptimisticLocking)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, persistence.ErrOptimisticLocking)
	assert.Nil(t, result)
}

func TestUpdateProductHandler_Handle_AttributeValidationFailure(t *testing.T) {
	repo, attrRepo, categoryRepo, _, _, _, handler := setupUpdateProductHandler(t)

	ctx := testCtxUpdate()
	existingProduct := createTestProduct()
	categoryID := "category-123"

	cmd := UpdateProductCommand{
		ID:         existingProduct.ID,
		Version:    existingProduct.Version,
		Name:       "Updated Product",
		Price:      199.99,
		Quantity:   20,
		ImageID:    ptr("image-123"),
		CategoryID: &categoryID,
		Enabled:    true,
		Attributes: []product.AttributeValue{
			{AttributeID: "non-existent-attr"},
		},
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingProduct.ID).
		Return(existingProduct, nil)

	categoryRepo.EXPECT().
		Exists(mock.Anything, categoryID).
		Return(true, nil)

	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{"non-existent-attr"}).
		Return(nil, errors.New("attribute not found"))

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Nil(t, result)
}
