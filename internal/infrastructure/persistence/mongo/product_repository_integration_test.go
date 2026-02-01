//go:build integration

package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductRepository_Insert(t *testing.T) {
	cleanupCollection(t, "product")

	ctx := context.Background()

	categoryID := uuid.New().String()
	imageID := uuid.New().String()
	prod, err := product.NewProduct(
		"Test Product",
		ptrI("A test product description"),
		99.99,
		10,
		&imageID,
		&categoryID,
		true,
		[]product.AttributeValue{
			{AttributeID: uuid.New().String(), OptionSlugValue: ptrI("red")},
			{AttributeID: uuid.New().String(), NumericValue: ptrI(float32(42))},
		},
	)
	require.NoError(t, err)

	err = testProductRepo.Insert(ctx, prod)
	require.NoError(t, err)

	// Verify by finding
	found, err := testProductRepo.FindByID(ctx, prod.ID)
	require.NoError(t, err)
	assert.Equal(t, prod.ID, found.ID)
	assert.Equal(t, prod.Name, found.Name)
	assert.Equal(t, *prod.Description, *found.Description)
	assert.Equal(t, prod.Price, found.Price)
	assert.True(t, found.Enabled)
	assert.Len(t, found.Attributes, 2)
}

func TestProductRepository_Update(t *testing.T) {
	cleanupCollection(t, "product")

	ctx := context.Background()

	prod, err := product.NewProduct(
		"Original Name",
		nil,
		10.00,
		5,
		nil,
		nil,
		false,
		nil,
	)
	require.NoError(t, err)

	err = testProductRepo.Insert(ctx, prod)
	require.NoError(t, err)

	// Update using domain method (modifies in place) - enable product requires image and category
	imageID := uuid.New().String()
	categoryID := uuid.New().String()
	err = prod.Update("Updated Name", ptrI("New description"), 20.00, 15, &imageID, &categoryID, true, nil)
	require.NoError(t, err)

	result, err := testProductRepo.Update(ctx, prod)
	require.NoError(t, err)

	// Verify
	found, err := testProductRepo.FindByID(ctx, prod.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, float32(20.00), found.Price)
	assert.True(t, found.Enabled)
	assert.Equal(t, result.Version, found.Version)
}

func TestProductRepository_FindByID(t *testing.T) {
	cleanupCollection(t, "product")

	ctx := context.Background()

	prod, err := product.NewProduct(
		"Find Me",
		nil,
		5.00,
		1,
		nil,
		nil,
		false,
		nil,
	)
	require.NoError(t, err)

	err = testProductRepo.Insert(ctx, prod)
	require.NoError(t, err)

	// Find existing
	found, err := testProductRepo.FindByID(ctx, prod.ID)
	require.NoError(t, err)
	assert.Equal(t, prod.ID, found.ID)

	// Find non-existing
	_, err = testProductRepo.FindByID(ctx, uuid.New().String())
	require.Error(t, err)
	assert.ErrorIs(t, err, persistence.ErrEntityNotFound)
}

func TestProductRepository_FindList(t *testing.T) {
	cleanupCollection(t, "product")

	ctx := context.Background()

	categoryID := uuid.New().String()
	imageID := uuid.New().String()

	// Create test products
	prod1, _ := product.NewProduct("Product 1", nil, 10.00, 1, nil, nil, false, nil)
	prod2, _ := product.NewProduct("Product 2", nil, 20.00, 2, &imageID, &categoryID, true, nil)
	prod3, _ := product.NewProduct("Product 3", nil, 30.00, 3, &imageID, &categoryID, true, nil)

	// Add delay to ensure different createdAt times
	require.NoError(t, testProductRepo.Insert(ctx, prod1))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, testProductRepo.Insert(ctx, prod2))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, testProductRepo.Insert(ctx, prod3))

	// List all
	query := product.ListQuery{
		Page: 1,
		Size: 10,
	}
	result, err := testProductRepo.FindList(ctx, query)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result.Total)
	assert.Len(t, result.Items, 3)

	// Filter by category
	queryWithCategory := product.ListQuery{
		CategoryID: &categoryID,
		Page:       1,
		Size:       10,
	}
	result, err = testProductRepo.FindList(ctx, queryWithCategory)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.Items, 2)

	// Filter by enabled
	enabled := true
	queryEnabled := product.ListQuery{
		Enabled: &enabled,
		Page:    1,
		Size:    10,
	}
	result, err = testProductRepo.FindList(ctx, queryEnabled)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)

	// Test pagination
	queryPaged := product.ListQuery{
		Page: 2,
		Size: 2,
	}
	result, err = testProductRepo.FindList(ctx, queryPaged)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result.Total)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, 2, result.Page)
}
