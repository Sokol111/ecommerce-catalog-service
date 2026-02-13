package category

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
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	categorymocks "github.com/Sokol111/ecommerce-catalog-service/internal/domain/category/mocks"
	eventmocks "github.com/Sokol111/ecommerce-catalog-service/internal/event/mocks"
	"github.com/Sokol111/ecommerce-catalog-service/internal/testutil/mocks"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

// createTestCategory creates a test category for update tests
func createTestCategory() *category.Category {
	return category.Reconstruct(
		"category-123",
		1,
		"Original Category",
		true,
		[]category.CategoryAttribute{
			{
				AttributeID: "attr-1",
				Slug:        "color",
				Role:        category.AttributeRoleVariant,
				Required:    true,
				SortOrder:   1,
				Filterable:  true,
				Searchable:  true,
			},
		},
		time.Now().UTC(),
		time.Now().UTC(),
	)
}

// setupUpdateCategoryHandler creates handler with mocked dependencies
func setupUpdateCategoryHandler(t *testing.T) (
	*categorymocks.MockRepository,
	*attributemocks.MockRepository,
	*mocks.MockOutbox,
	*mocks.MockTxManager,
	*eventmocks.MockCategoryEventFactory,
	UpdateCategoryCommandHandler,
) {
	repo := categorymocks.NewMockRepository(t)
	attrRepo := attributemocks.NewMockRepository(t)
	outboxMock := mocks.NewMockOutbox(t)
	txManager := mocks.NewMockTxManager(t)
	eventFactory := eventmocks.NewMockCategoryEventFactory(t)

	handler := NewUpdateCategoryHandler(repo, attrRepo, outboxMock, txManager, eventFactory)

	return repo, attrRepo, outboxMock, txManager, eventFactory, handler
}

func TestUpdateCategoryHandler_Handle_Success(t *testing.T) {
	repo, attrRepo, outboxMock, txManager, eventFactory, handler := setupUpdateCategoryHandler(t)

	ctx := testCtx()
	existingCategory := createTestCategory()

	cmd := UpdateCategoryCommand{
		ID:      existingCategory.ID,
		Version: existingCategory.Version,
		Name:    "Updated Category",
		Enabled: false,
		Attributes: []CategoryAttributeInput{
			{
				AttributeID: "attr-2",
				Role:        "specification",
				Required:    false,
				SortOrder:   1,
				Filterable:  true,
				Searchable:  false,
			},
		},
	}

	// Mock find existing category
	repo.EXPECT().
		FindByID(mock.Anything, existingCategory.ID).
		Return(existingCategory, nil)

	// Mock attribute lookup
	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{"attr-2"}).
		Return([]*attribute.Attribute{
			attribute.Reconstruct("attr-2", 1, "Size", "size", attribute.AttributeTypeSingle, nil, true, nil, time.Now(), time.Now()),
		}, nil)

	// Mock transaction
	txManager.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	// Mock update
	repo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*category.Category")).
		RunAndReturn(func(_ context.Context, c *category.Category) (*category.Category, error) {
			return c, nil
		})

	// Mock event factory
	eventFactory.EXPECT().
		NewCategoryUpdatedOutboxMessage(mock.Anything, mock.Anything).
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
	assert.False(t, result.Enabled)
	assert.Len(t, result.Attributes, 1)
	assert.Equal(t, "size", result.Attributes[0].Slug)
}

func TestUpdateCategoryHandler_Handle_NotFound(t *testing.T) {
	repo, _, _, _, _, handler := setupUpdateCategoryHandler(t)

	ctx := testCtx()
	cmd := UpdateCategoryCommand{
		ID:      "non-existent-id",
		Version: 1,
		Name:    "Updated Category",
	}

	repo.EXPECT().
		FindByID(mock.Anything, cmd.ID).
		Return(nil, mongo.ErrEntityNotFound)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, mongo.ErrEntityNotFound)
	assert.Nil(t, result)
}

func TestUpdateCategoryHandler_Handle_OptimisticLockingVersionMismatch(t *testing.T) {
	repo, _, _, _, _, handler := setupUpdateCategoryHandler(t)

	ctx := testCtx()
	existingCategory := createTestCategory() // Version 1

	cmd := UpdateCategoryCommand{
		ID:      existingCategory.ID,
		Version: 2, // Wrong version
		Name:    "Updated Category",
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingCategory.ID).
		Return(existingCategory, nil)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, mongo.ErrOptimisticLocking)
	assert.Nil(t, result)
}

func TestUpdateCategoryHandler_Handle_InvalidName(t *testing.T) {
	repo, attrRepo, _, _, _, handler := setupUpdateCategoryHandler(t)

	ctx := testCtx()
	existingCategory := createTestCategory()

	cmd := UpdateCategoryCommand{
		ID:      existingCategory.ID,
		Version: existingCategory.Version,
		Name:    "", // Invalid - empty name
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingCategory.ID).
		Return(existingCategory, nil)

	// Mock empty attributes lookup
	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{}).
		Return([]*attribute.Attribute{}, nil)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update category")
	assert.Nil(t, result)
}

func TestUpdateCategoryHandler_Handle_AttributeNotFound(t *testing.T) {
	repo, attrRepo, _, _, _, handler := setupUpdateCategoryHandler(t)

	ctx := testCtx()
	existingCategory := createTestCategory()

	cmd := UpdateCategoryCommand{
		ID:      existingCategory.ID,
		Version: existingCategory.Version,
		Name:    "Updated Category",
		Attributes: []CategoryAttributeInput{
			{AttributeID: "non-existent-attr"},
		},
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingCategory.ID).
		Return(existingCategory, nil)

	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{"non-existent-attr"}).
		Return(nil, errors.New("attribute not found"))

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestUpdateCategoryHandler_Handle_UpdateRepositoryError(t *testing.T) {
	repo, attrRepo, _, txManager, _, handler := setupUpdateCategoryHandler(t)

	ctx := testCtx()
	existingCategory := createTestCategory()

	cmd := UpdateCategoryCommand{
		ID:      existingCategory.ID,
		Version: existingCategory.Version,
		Name:    "Updated Category",
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingCategory.ID).
		Return(existingCategory, nil)

	// Mock empty attributes lookup
	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{}).
		Return([]*attribute.Attribute{}, nil)

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
	assert.Contains(t, err.Error(), "failed to update category")
	assert.Nil(t, result)
}

func TestUpdateCategoryHandler_Handle_OptimisticLockingOnUpdate(t *testing.T) {
	repo, attrRepo, _, txManager, _, handler := setupUpdateCategoryHandler(t)

	ctx := testCtx()
	existingCategory := createTestCategory()

	cmd := UpdateCategoryCommand{
		ID:      existingCategory.ID,
		Version: existingCategory.Version,
		Name:    "Updated Category",
	}

	repo.EXPECT().
		FindByID(mock.Anything, existingCategory.ID).
		Return(existingCategory, nil)

	// Mock empty attributes lookup
	attrRepo.EXPECT().
		FindByIDsOrFail(mock.Anything, []string{}).
		Return([]*attribute.Attribute{}, nil)

	txManager.EXPECT().
		WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	repo.EXPECT().
		Update(mock.Anything, mock.Anything).
		Return(nil, mongo.ErrOptimisticLocking)

	result, err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, mongo.ErrOptimisticLocking)
	assert.Nil(t, result)
}
