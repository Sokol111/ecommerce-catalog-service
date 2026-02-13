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

	list, ok := listResp.(*httpapi.AttributeListResponse)
	require.True(t, ok, "expected *httpapi.AttributeListResponse, got %T", listResp)
	assert.GreaterOrEqual(t, len(list.Items), 3)
	assert.GreaterOrEqual(t, list.Total, 3)
}

func TestAttribute_ListWithFilters(t *testing.T) {
	ctx := context.Background()
	uniqueSuffix := uuid.New().String()[:8]

	// 1. Create attributes with different types and enabled status
	enabledSingle := &httpapi.CreateAttributeRequest{
		Name:    "FilterTest Enabled Single",
		Slug:    "filter-enabled-single-" + uniqueSuffix,
		Type:    httpapi.CreateAttributeRequestTypeSingle,
		Enabled: true,
		Options: []httpapi.AttributeOptionInput{
			{Name: "Option1", Slug: "opt1", SortOrder: httpapi.NewOptInt(1)},
		},
	}
	_, err := testClient.CreateAttribute(ctx, enabledSingle)
	require.NoError(t, err)

	disabledBoolean := &httpapi.CreateAttributeRequest{
		Name:    "FilterTest Disabled Boolean",
		Slug:    "filter-disabled-bool-" + uniqueSuffix,
		Type:    httpapi.CreateAttributeRequestTypeBoolean,
		Enabled: false,
	}
	_, err = testClient.CreateAttribute(ctx, disabledBoolean)
	require.NoError(t, err)

	// 2. Test filter by enabled=true
	listResp, err := testClient.GetAttributeList(ctx, httpapi.GetAttributeListParams{
		Page:    1,
		Size:    100,
		Enabled: httpapi.NewOptBool(true),
	})
	require.NoError(t, err)
	list := listResp.(*httpapi.AttributeListResponse)

	for _, item := range list.Items {
		assert.True(t, item.Enabled, "expected all items to be enabled")
	}

	// 3. Test filter by type=boolean
	listResp, err = testClient.GetAttributeList(ctx, httpapi.GetAttributeListParams{
		Page: 1,
		Size: 100,
		Type: httpapi.NewOptGetAttributeListType(httpapi.GetAttributeListTypeBoolean),
	})
	require.NoError(t, err)
	list = listResp.(*httpapi.AttributeListResponse)

	for _, item := range list.Items {
		assert.Equal(t, httpapi.AttributeResponseTypeBoolean, item.Type, "expected all items to be boolean type")
	}

	// 4. Test sorting by name desc
	listResp, err = testClient.GetAttributeList(ctx, httpapi.GetAttributeListParams{
		Page:  1,
		Size:  10,
		Sort:  httpapi.NewOptGetAttributeListSort(httpapi.GetAttributeListSortName),
		Order: httpapi.NewOptGetAttributeListOrder(httpapi.GetAttributeListOrderDesc),
	})
	require.NoError(t, err)
	list = listResp.(*httpapi.AttributeListResponse)

	// Verify descending order
	for i := 1; i < len(list.Items); i++ {
		assert.GreaterOrEqual(t, list.Items[i-1].Name, list.Items[i].Name,
			"expected items sorted by name descending")
	}
}

func TestAttribute_NotFound(t *testing.T) {
	ctx := context.Background()

	// Try to get non-existent attribute
	resp, err := testClient.GetAttributeById(ctx, httpapi.GetAttributeByIdParams{
		ID: uuid.New(),
	})
	require.NoError(t, err)

	// Should return NotFound response type
	_, ok := resp.(*httpapi.GetAttributeByIdNotFound)
	assert.True(t, ok, "expected *httpapi.GetAttributeByIdNotFound, got %T", resp)
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
	resp, err := testClient.CreateAttribute(ctx, duplicateReq)
	require.NoError(t, err)

	// Should return Conflict response type
	_, ok := resp.(*httpapi.CreateAttributeConflict)
	assert.True(t, ok, "expected *httpapi.CreateAttributeConflict, got %T", resp)
}

func TestAttribute_UpdateNotFound(t *testing.T) {
	ctx := context.Background()

	// Try to update non-existent attribute
	updateReq := &httpapi.UpdateAttributeRequest{
		ID:      uuid.New(),
		Name:    "Non-existent",
		Enabled: true,
		Version: 1,
	}

	resp, err := testClient.UpdateAttribute(ctx, updateReq)
	require.NoError(t, err)

	// Should return NotFound response type
	_, ok := resp.(*httpapi.UpdateAttributeNotFound)
	assert.True(t, ok, "expected *httpapi.UpdateAttributeNotFound, got %T", resp)
}

func TestAttribute_UpdateVersionMismatch(t *testing.T) {
	ctx := context.Background()

	// 1. Create attribute
	createReq := &httpapi.CreateAttributeRequest{
		Name:    "Version Test",
		Slug:    "version-test-" + uuid.New().String()[:8],
		Type:    httpapi.CreateAttributeRequestTypeText,
		Enabled: true,
	}

	createResp, err := testClient.CreateAttribute(ctx, createReq)
	require.NoError(t, err)
	created := createResp.(*httpapi.AttributeResponse)

	createdID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	// 2. Try to update with wrong version
	updateReq := &httpapi.UpdateAttributeRequest{
		ID:      createdID,
		Name:    "Updated Name",
		Enabled: true,
		Version: 999, // Wrong version
	}

	resp, err := testClient.UpdateAttribute(ctx, updateReq)
	require.NoError(t, err)

	// Should return PreconditionFailed response type
	_, ok := resp.(*httpapi.UpdateAttributePreconditionFailed)
	assert.True(t, ok, "expected *httpapi.UpdateAttributePreconditionFailed, got %T", resp)
}
