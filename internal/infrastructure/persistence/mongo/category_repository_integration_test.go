//go:build integration

package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoryRepository_Insert(t *testing.T) {
	cleanupCollection(t, "category")

	ctx := context.Background()

	cat, err := category.NewCategory(
		"Electronics",
		true,
		[]category.CategoryAttribute{
			{AttributeID: uuid.New().String(), Slug: "color", Role: category.AttributeRoleVariant, Required: true, SortOrder: 1, Filterable: true},
			{AttributeID: uuid.New().String(), Slug: "size", Role: category.AttributeRoleSpecification, Required: false, SortOrder: 2, Filterable: false},
		},
	)
	require.NoError(t, err)

	err = testCategoryRepo.Insert(ctx, cat)
	require.NoError(t, err)

	// Verify by finding
	found, err := testCategoryRepo.FindByID(ctx, cat.ID)
	require.NoError(t, err)
	assert.Equal(t, cat.ID, found.ID)
	assert.Equal(t, cat.Name, found.Name)
	assert.True(t, found.Enabled)
	assert.Len(t, found.Attributes, 2)
}

func TestCategoryRepository_Update(t *testing.T) {
	cleanupCollection(t, "category")

	ctx := context.Background()

	cat, err := category.NewCategory(
		"Clothing",
		true,
		nil,
	)
	require.NoError(t, err)

	err = testCategoryRepo.Insert(ctx, cat)
	require.NoError(t, err)

	// Update using domain method (modifies in place)
	err = cat.Update("Apparel", false, nil)
	require.NoError(t, err)

	result, err := testCategoryRepo.Update(ctx, cat)
	require.NoError(t, err)

	// Verify
	found, err := testCategoryRepo.FindByID(ctx, cat.ID)
	require.NoError(t, err)
	assert.Equal(t, "Apparel", found.Name)
	assert.False(t, found.Enabled)
	assert.Equal(t, result.Version, found.Version)
}

func TestCategoryRepository_FindByID(t *testing.T) {
	cleanupCollection(t, "category")

	ctx := context.Background()

	cat, err := category.NewCategory(
		"Books",
		true,
		nil,
	)
	require.NoError(t, err)

	err = testCategoryRepo.Insert(ctx, cat)
	require.NoError(t, err)

	// Find existing
	found, err := testCategoryRepo.FindByID(ctx, cat.ID)
	require.NoError(t, err)
	assert.Equal(t, cat.ID, found.ID)

	// Find non-existing
	_, err = testCategoryRepo.FindByID(ctx, uuid.New().String())
	require.Error(t, err)
	assert.ErrorIs(t, err, persistence.ErrEntityNotFound)
}

func TestCategoryRepository_FindList(t *testing.T) {
	cleanupCollection(t, "category")

	ctx := context.Background()

	// Create test categories
	cat1, _ := category.NewCategory("Category 1", true, nil)
	cat2, _ := category.NewCategory("Category 2", true, nil)
	cat3, _ := category.NewCategory("Category 3", false, nil)

	// Add delay to ensure different createdAt times
	require.NoError(t, testCategoryRepo.Insert(ctx, cat1))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, testCategoryRepo.Insert(ctx, cat2))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, testCategoryRepo.Insert(ctx, cat3))

	// List all
	query := category.ListQuery{
		Page: 1,
		Size: 10,
	}
	result, err := testCategoryRepo.FindList(ctx, query)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result.Total)
	assert.Len(t, result.Items, 3)

	// Filter by enabled
	enabled := true
	queryEnabled := category.ListQuery{
		Enabled: &enabled,
		Page:    1,
		Size:    10,
	}
	result, err = testCategoryRepo.FindList(ctx, queryEnabled)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)

	// Test pagination
	queryPaged := category.ListQuery{
		Page: 2,
		Size: 2,
	}
	result, err = testCategoryRepo.FindList(ctx, queryPaged)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result.Total)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, 2, result.Page)
}

func TestCategoryRepository_Exists(t *testing.T) {
	cleanupCollection(t, "category")

	ctx := context.Background()

	cat, _ := category.NewCategory("Test Category", true, nil)

	// Should not exist initially
	exists, err := testCategoryRepo.Exists(ctx, cat.ID)
	require.NoError(t, err)
	assert.False(t, exists)

	// Create
	err = testCategoryRepo.Insert(ctx, cat)
	require.NoError(t, err)

	// Should exist now
	exists, err = testCategoryRepo.Exists(ctx, cat.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Different ID should not exist
	exists, err = testCategoryRepo.Exists(ctx, uuid.New().String())
	require.NoError(t, err)
	assert.False(t, exists)
}
