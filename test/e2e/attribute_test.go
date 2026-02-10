//go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/httpapi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttribute_CreateAndGet(t *testing.T) {
	ctx := context.Background()

	// 1. Create attribute via API
	createReq := &httpapi.CreateAttributeRequest{
		Name:    "Color",
		Slug:    "color-e2e-test-" + uuid.New().String()[:8],
		Type:    httpapi.CreateAttributeRequestTypeSingle,
		Enabled: true,
		Options: []httpapi.AttributeOptionInput{
			{Name: "Red", Slug: "red", SortOrder: httpapi.NewOptInt(1)},
			{Name: "Blue", Slug: "blue", SortOrder: httpapi.NewOptInt(2)},
		},
	}

	createResp, err := testClient.CreateAttribute(ctx, createReq)
	require.NoError(t, err)
	require.IsType(t, &httpapi.AttributeResponse{}, createResp)

	created := createResp.(*httpapi.AttributeResponse)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "Color", created.Name)
	assert.Equal(t, httpapi.AttributeResponseTypeSingle, created.Type)
	assert.True(t, created.Enabled)
	assert.Len(t, created.Options, 2)

	// 2. Get attribute by ID
	createdID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	getResp, err := testClient.GetAttributeById(ctx, httpapi.GetAttributeByIdParams{
		ID: createdID,
	})
	require.NoError(t, err)
	require.IsType(t, &httpapi.AttributeResponse{}, getResp)

	fetched := getResp.(*httpapi.AttributeResponse)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, created.Name, fetched.Name)
	assert.Equal(t, created.Slug, fetched.Slug)
}

func TestAttribute_Update(t *testing.T) {
	ctx := context.Background()

	// 1. Create attribute
	createReq := &httpapi.CreateAttributeRequest{
		Name:    "Size",
		Slug:    "size-e2e-update-" + uuid.New().String()[:8],
		Type:    httpapi.CreateAttributeRequestTypeSingle,
		Enabled: false,
	}

	createResp, err := testClient.CreateAttribute(ctx, createReq)
	require.NoError(t, err)
	created := createResp.(*httpapi.AttributeResponse)

	// 2. Update attribute
	createdID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	updateReq := &httpapi.UpdateAttributeRequest{
		ID:      createdID,
		Name:    "Size Updated",
		Enabled: true,
		Version: created.Version,
		Options: []httpapi.AttributeOptionInput{
			{Name: "Small", Slug: "small", SortOrder: httpapi.NewOptInt(1)},
			{Name: "Large", Slug: "large", SortOrder: httpapi.NewOptInt(2)},
		},
	}

	updateResp, err := testClient.UpdateAttribute(ctx, updateReq)
	require.NoError(t, err)
	require.IsType(t, &httpapi.AttributeResponse{}, updateResp)

	updated := updateResp.(*httpapi.AttributeResponse)
	assert.Equal(t, "Size Updated", updated.Name)
	assert.True(t, updated.Enabled)
	assert.Len(t, updated.Options, 2)
	assert.Equal(t, 2, updated.Version) // Version incremented
}

func TestAttribute_List(t *testing.T) {
	ctx := context.Background()

	// 1. Create multiple attributes
	for i := 1; i <= 3; i++ {
		createReq := &httpapi.CreateAttributeRequest{
			Name:    "ListTest" + string(rune('0'+i)),
			Slug:    "list-test-" + uuid.New().String()[:8],
			Type:    httpapi.CreateAttributeRequestTypeText,
			Enabled: true,
		}
		_, err := testClient.CreateAttribute(ctx, createReq)
		require.NoError(t, err)
	}

	// 2. Get list
	listResp, err := testClient.GetAttributeList(ctx, httpapi.GetAttributeListParams{
		Page: 1,
		Size: 10,
	})
	require.NoError(t, err)
	require.IsType(t, &httpapi.AttributeListResponse{}, listResp)

	list := listResp.(*httpapi.AttributeListResponse)
	assert.GreaterOrEqual(t, len(list.Items), 3)
	assert.GreaterOrEqual(t, list.Total, int64(3))
}

func TestAttribute_NotFound(t *testing.T) {
	ctx := context.Background()

	// Try to get non-existent attribute
	_, err := testClient.GetAttributeById(ctx, httpapi.GetAttributeByIdParams{
		ID: uuid.New(),
	})

	// Should return error (404)
	require.Error(t, err)
}

func TestAttribute_DuplicateSlug(t *testing.T) {
	ctx := context.Background()

	uniqueSlug := "unique-slug-e2e-" + uuid.New().String()[:8]

	// 1. Create first attribute
	createReq := &httpapi.CreateAttributeRequest{
		Name:    "Unique Attr",
		Slug:    uniqueSlug,
		Type:    httpapi.CreateAttributeRequestTypeText,
		Enabled: true,
	}
	_, err := testClient.CreateAttribute(ctx, createReq)
	require.NoError(t, err)

	// 2. Try to create with same slug
	duplicateReq := &httpapi.CreateAttributeRequest{
		Name:    "Another Attr",
		Slug:    uniqueSlug, // Same slug
		Type:    httpapi.CreateAttributeRequestTypeBoolean,
		Enabled: true,
	}
	_, err = testClient.CreateAttribute(ctx, duplicateReq)

	// Should return error (409 Conflict)
	require.Error(t, err)
}
