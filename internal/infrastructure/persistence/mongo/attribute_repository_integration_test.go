//go:build integration

package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttributeRepository_Insert(t *testing.T) {
	cleanupCollection(t, "attribute")

	ctx := context.Background()

	attr, err := attribute.NewAttribute(
		uuid.New().String(),
		"Color",
		"color",
		attribute.AttributeTypeSingle,
		nil,
		true,
		[]attribute.Option{
			{Name: "Red", Slug: "red", SortOrder: 1},
			{Name: "Blue", Slug: "blue", SortOrder: 2},
		},
	)
	require.NoError(t, err)

	err = testAttributeRepo.Insert(ctx, attr)
	require.NoError(t, err)

	// Verify by finding
	found, err := testAttributeRepo.FindByID(ctx, attr.ID)
	require.NoError(t, err)
	assert.Equal(t, attr.ID, found.ID)
	assert.Equal(t, attr.Slug, found.Slug)
	assert.Equal(t, attr.Name, found.Name)
	assert.Equal(t, attr.Type, found.Type)
	assert.Len(t, found.Options, 2)
}

func TestAttributeRepository_Insert_DuplicateSlug(t *testing.T) {
	cleanupCollection(t, "attribute")

	ctx := context.Background()

	attr1, err := attribute.NewAttribute(
		uuid.New().String(),
		"Size",
		"size",
		attribute.AttributeTypeSingle,
		nil,
		true,
		nil,
	)
	require.NoError(t, err)

	attr2, err := attribute.NewAttribute(
		uuid.New().String(),
		"Size 2",
		"size", // Same slug
		attribute.AttributeTypeMultiple,
		nil,
		true,
		nil,
	)
	require.NoError(t, err)

	err = testAttributeRepo.Insert(ctx, attr1)
	require.NoError(t, err)

	err = testAttributeRepo.Insert(ctx, attr2)
	require.Error(t, err)
	// Repository converts MongoDB duplicate key error to domain error
	assert.ErrorIs(t, err, attribute.ErrSlugAlreadyExists)
}

func TestAttributeRepository_Update(t *testing.T) {
	cleanupCollection(t, "attribute")

	ctx := context.Background()

	attr, err := attribute.NewAttribute(
		uuid.New().String(),
		"Material",
		"material",
		attribute.AttributeTypeText,
		nil,
		false,
		nil,
	)
	require.NoError(t, err)

	err = testAttributeRepo.Insert(ctx, attr)
	require.NoError(t, err)

	// Update using domain method (modifies in place)
	err = attr.Update("Material Type", ptrI("kg"), true, nil)
	require.NoError(t, err)

	result, err := testAttributeRepo.Update(ctx, attr)
	require.NoError(t, err)

	// Verify
	found, err := testAttributeRepo.FindByID(ctx, attr.ID)
	require.NoError(t, err)
	assert.Equal(t, "Material Type", found.Name)
	assert.True(t, found.Enabled)
	assert.Equal(t, result.Version, found.Version)
}

func TestAttributeRepository_FindByID(t *testing.T) {
	cleanupCollection(t, "attribute")

	ctx := context.Background()

	attr, err := attribute.NewAttribute(
		uuid.New().String(),
		"Weight",
		"weight",
		attribute.AttributeTypeRange,
		ptrI("kg"),
		true,
		nil,
	)
	require.NoError(t, err)

	err = testAttributeRepo.Insert(ctx, attr)
	require.NoError(t, err)

	// Find existing
	found, err := testAttributeRepo.FindByID(ctx, attr.ID)
	require.NoError(t, err)
	assert.Equal(t, attr.ID, found.ID)

	// Find non-existing
	_, err = testAttributeRepo.FindByID(ctx, uuid.New().String())
	require.Error(t, err)
	assert.ErrorIs(t, err, mongo.ErrEntityNotFound)
}

func TestAttributeRepository_FindByIDs(t *testing.T) {
	cleanupCollection(t, "attribute")

	ctx := context.Background()

	attr1, _ := attribute.NewAttribute(uuid.New().String(), "Attr1", "attr1", attribute.AttributeTypeText, nil, true, nil)
	attr2, _ := attribute.NewAttribute(uuid.New().String(), "Attr2", "attr2", attribute.AttributeTypeSingle, nil, true, nil)
	attr3, _ := attribute.NewAttribute(uuid.New().String(), "Attr3", "attr3", attribute.AttributeTypeBoolean, nil, true, nil)

	require.NoError(t, testAttributeRepo.Insert(ctx, attr1))
	require.NoError(t, testAttributeRepo.Insert(ctx, attr2))
	require.NoError(t, testAttributeRepo.Insert(ctx, attr3))

	// Find multiple
	ids := []string{attr1.ID, attr3.ID}
	found, err := testAttributeRepo.FindByIDs(ctx, ids)
	require.NoError(t, err)
	assert.Len(t, found, 2)

	// Find with non-existing ID (should return only found ones)
	idsWithNonExisting := []string{attr1.ID, uuid.New().String()}
	found, err = testAttributeRepo.FindByIDs(ctx, idsWithNonExisting)
	require.NoError(t, err)
	assert.Len(t, found, 1)
}

func TestAttributeRepository_FindByIDsOrFail(t *testing.T) {
	cleanupCollection(t, "attribute")

	ctx := context.Background()

	attr1, _ := attribute.NewAttribute(uuid.New().String(), "Attr1", "attr1", attribute.AttributeTypeText, nil, true, nil)
	attr2, _ := attribute.NewAttribute(uuid.New().String(), "Attr2", "attr2", attribute.AttributeTypeSingle, nil, true, nil)

	require.NoError(t, testAttributeRepo.Insert(ctx, attr1))
	require.NoError(t, testAttributeRepo.Insert(ctx, attr2))

	// Find all existing - should succeed
	ids := []string{attr1.ID, attr2.ID}
	found, err := testAttributeRepo.FindByIDsOrFail(ctx, ids)
	require.NoError(t, err)
	assert.Len(t, found, 2)

	// Find with non-existing ID - should fail
	idsWithNonExisting := []string{attr1.ID, uuid.New().String()}
	_, err = testAttributeRepo.FindByIDsOrFail(ctx, idsWithNonExisting)
	require.Error(t, err)
}

func TestAttributeRepository_FindList(t *testing.T) {
	cleanupCollection(t, "attribute")

	ctx := context.Background()

	// Create test attributes
	attr1, _ := attribute.NewAttribute(uuid.New().String(), "Attribute 1", "attr1", attribute.AttributeTypeText, nil, true, nil)
	attr2, _ := attribute.NewAttribute(uuid.New().String(), "Attribute 2", "attr2", attribute.AttributeTypeSingle, nil, true, nil)
	attr3, _ := attribute.NewAttribute(uuid.New().String(), "Attribute 3", "attr3", attribute.AttributeTypeRange, nil, false, nil)

	// Add delay to ensure different createdAt times
	require.NoError(t, testAttributeRepo.Insert(ctx, attr1))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, testAttributeRepo.Insert(ctx, attr2))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, testAttributeRepo.Insert(ctx, attr3))

	// List all
	query := attribute.ListQuery{
		Page: 1,
		Size: 10,
	}
	result, err := testAttributeRepo.FindList(ctx, query)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result.Total)
	assert.Len(t, result.Items, 3)

	// Filter by type
	singleType := string(attribute.AttributeTypeSingle)
	queryWithType := attribute.ListQuery{
		Type: &singleType,
		Page: 1,
		Size: 10,
	}
	result, err = testAttributeRepo.FindList(ctx, queryWithType)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, attribute.AttributeTypeSingle, result.Items[0].Type)

	// Filter by enabled
	enabled := true
	queryEnabled := attribute.ListQuery{
		Enabled: &enabled,
		Page:    1,
		Size:    10,
	}
	result, err = testAttributeRepo.FindList(ctx, queryEnabled)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)

	// Test pagination
	queryPaged := attribute.ListQuery{
		Page: 2,
		Size: 2,
	}
	result, err = testAttributeRepo.FindList(ctx, queryPaged)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result.Total)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, 2, result.Page)
}

func TestAttributeRepository_Exists(t *testing.T) {
	cleanupCollection(t, "attribute")

	ctx := context.Background()

	attr, _ := attribute.NewAttribute(uuid.New().String(), "Unique", "unique-slug", attribute.AttributeTypeText, nil, true, nil)

	// Should not exist initially
	exists, err := testAttributeRepo.Exists(ctx, attr.ID)
	require.NoError(t, err)
	assert.False(t, exists)

	// Create
	err = testAttributeRepo.Insert(ctx, attr)
	require.NoError(t, err)

	// Should exist now
	exists, err = testAttributeRepo.Exists(ctx, attr.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Different ID should not exist
	exists, err = testAttributeRepo.Exists(ctx, uuid.New().String())
	require.NoError(t, err)
	assert.False(t, exists)
}
