//go:build e2e

package e2e

import (
	"context"
	"testing"

	catalogv1 "github.com/Sokol111/ecommerce-catalog-service-api/gen/go/catalog/v1"
	"github.com/Sokol111/ecommerce-commons/pkg/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func testContext() context.Context {
	return tenant.ContextWithSlug(context.Background(), "default")
}

func TestAttribute_CreateAndGet(t *testing.T) {
	ctx := testContext()

	createResp, err := testAttributeClient.CreateAttribute(ctx, &catalogv1.CreateAttributeRequest{
		Name:    "Color",
		Slug:    "color-e2e-test-" + uuid.New().String()[:8],
		Type:    catalogv1.AttributeType_ATTRIBUTE_TYPE_SINGLE,
		Enabled: true,
		Options: []*catalogv1.AttributeOptionInput{
			{Name: "Red", Slug: "red", SortOrder: int32Ptr(1)},
			{Name: "Blue", Slug: "blue", SortOrder: int32Ptr(2)},
		},
	})
	require.NoError(t, err)

	created := createResp.GetAttribute()
	assert.NotEmpty(t, created.GetId())
	assert.Equal(t, "Color", created.GetName())
	assert.Equal(t, catalogv1.AttributeType_ATTRIBUTE_TYPE_SINGLE, created.GetType())
	assert.True(t, created.GetEnabled())
	assert.Len(t, created.GetOptions(), 2)

	getResp, err := testAttributeClient.GetAttributeById(ctx, &catalogv1.GetAttributeByIdRequest{
		Id: created.GetId(),
	})
	require.NoError(t, err)

	fetched := getResp.GetAttribute()
	assert.Equal(t, created.GetId(), fetched.GetId())
	assert.Equal(t, created.GetName(), fetched.GetName())
	assert.Equal(t, created.GetSlug(), fetched.GetSlug())
}

func TestAttribute_Update(t *testing.T) {
	ctx := testContext()

	createResp, err := testAttributeClient.CreateAttribute(ctx, &catalogv1.CreateAttributeRequest{
		Name:    "Size",
		Slug:    "size-e2e-update-" + uuid.New().String()[:8],
		Type:    catalogv1.AttributeType_ATTRIBUTE_TYPE_SINGLE,
		Enabled: false,
	})
	require.NoError(t, err)
	created := createResp.GetAttribute()

	updateResp, err := testAttributeClient.UpdateAttribute(ctx, &catalogv1.UpdateAttributeRequest{
		Id:      created.GetId(),
		Name:    "Size Updated",
		Enabled: true,
		Version: created.GetVersion(),
		Options: []*catalogv1.AttributeOptionInput{
			{Name: "Small", Slug: "small", SortOrder: int32Ptr(1)},
			{Name: "Large", Slug: "large", SortOrder: int32Ptr(2)},
		},
	})
	require.NoError(t, err)

	updated := updateResp.GetAttribute()
	assert.Equal(t, "Size Updated", updated.GetName())
	assert.True(t, updated.GetEnabled())
	assert.Len(t, updated.GetOptions(), 2)
	assert.Equal(t, int64(2), updated.GetVersion())
}

func TestAttribute_List(t *testing.T) {
	ctx := testContext()

	for i := 1; i <= 3; i++ {
		_, err := testAttributeClient.CreateAttribute(ctx, &catalogv1.CreateAttributeRequest{
			Name:    "ListTest" + string(rune('0'+i)),
			Slug:    "list-test-" + uuid.New().String()[:8],
			Type:    catalogv1.AttributeType_ATTRIBUTE_TYPE_TEXT,
			Enabled: true,
		})
		require.NoError(t, err)
	}

	listResp, err := testAttributeClient.GetAttributeList(ctx, &catalogv1.GetAttributeListRequest{
		Page: 1,
		Size: 10,
	})
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(listResp.GetItems()), 3)
	assert.GreaterOrEqual(t, listResp.GetTotal(), int64(3))
}

func TestAttribute_ListWithFilters(t *testing.T) {
	ctx := testContext()
	uniqueSuffix := uuid.New().String()[:8]

	_, err := testAttributeClient.CreateAttribute(ctx, &catalogv1.CreateAttributeRequest{
		Name:    "FilterTest Enabled Single",
		Slug:    "filter-enabled-single-" + uniqueSuffix,
		Type:    catalogv1.AttributeType_ATTRIBUTE_TYPE_SINGLE,
		Enabled: true,
		Options: []*catalogv1.AttributeOptionInput{
			{Name: "Option1", Slug: "opt1", SortOrder: int32Ptr(1)},
		},
	})
	require.NoError(t, err)

	_, err = testAttributeClient.CreateAttribute(ctx, &catalogv1.CreateAttributeRequest{
		Name:    "FilterTest Disabled Boolean",
		Slug:    "filter-disabled-bool-" + uniqueSuffix,
		Type:    catalogv1.AttributeType_ATTRIBUTE_TYPE_BOOLEAN,
		Enabled: false,
	})
	require.NoError(t, err)

	enabled := true
	listResp, err := testAttributeClient.GetAttributeList(ctx, &catalogv1.GetAttributeListRequest{
		Page:    1,
		Size:    100,
		Enabled: &enabled,
	})
	require.NoError(t, err)
	for _, item := range listResp.GetItems() {
		assert.True(t, item.GetEnabled(), "expected all items to be enabled")
	}

	boolType := catalogv1.AttributeType_ATTRIBUTE_TYPE_BOOLEAN
	listResp, err = testAttributeClient.GetAttributeList(ctx, &catalogv1.GetAttributeListRequest{
		Page: 1,
		Size: 100,
		Type: &boolType,
	})
	require.NoError(t, err)
	for _, item := range listResp.GetItems() {
		assert.Equal(t, catalogv1.AttributeType_ATTRIBUTE_TYPE_BOOLEAN, item.GetType(), "expected all items to be boolean type")
	}

	sortName := "name"
	sortDesc := "desc"
	listResp, err = testAttributeClient.GetAttributeList(ctx, &catalogv1.GetAttributeListRequest{
		Page:  1,
		Size:  10,
		Sort:  &sortName,
		Order: &sortDesc,
	})
	require.NoError(t, err)
	items := listResp.GetItems()
	for i := 1; i < len(items); i++ {
		assert.GreaterOrEqual(t, items[i-1].GetName(), items[i].GetName(),
			"expected items sorted by name descending")
	}
}

func TestAttribute_NotFound(t *testing.T) {
	ctx := testContext()

	_, err := testAttributeClient.GetAttributeById(ctx, &catalogv1.GetAttributeByIdRequest{
		Id: uuid.New().String(),
	})
	require.Error(t, err)

	statusErr, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, statusErr.Code())
}

func TestAttribute_DuplicateSlug(t *testing.T) {
	ctx := testContext()
	uniqueSlug := "unique-slug-e2e-" + uuid.New().String()[:8]

	_, err := testAttributeClient.CreateAttribute(ctx, &catalogv1.CreateAttributeRequest{
		Name:    "Unique Attr",
		Slug:    uniqueSlug,
		Type:    catalogv1.AttributeType_ATTRIBUTE_TYPE_TEXT,
		Enabled: true,
	})
	require.NoError(t, err)

	_, err = testAttributeClient.CreateAttribute(ctx, &catalogv1.CreateAttributeRequest{
		Name:    "Another Attr",
		Slug:    uniqueSlug,
		Type:    catalogv1.AttributeType_ATTRIBUTE_TYPE_BOOLEAN,
		Enabled: true,
	})
	require.Error(t, err)

	statusErr, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, statusErr.Code())
}

func TestAttribute_UpdateNotFound(t *testing.T) {
	ctx := testContext()

	_, err := testAttributeClient.UpdateAttribute(ctx, &catalogv1.UpdateAttributeRequest{
		Id:      uuid.New().String(),
		Name:    "Non-existent",
		Enabled: true,
		Version: 1,
	})
	require.Error(t, err)

	statusErr, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, statusErr.Code())
}

func TestAttribute_UpdateVersionMismatch(t *testing.T) {
	ctx := testContext()

	createResp, err := testAttributeClient.CreateAttribute(ctx, &catalogv1.CreateAttributeRequest{
		Name:    "Version Test",
		Slug:    "version-test-" + uuid.New().String()[:8],
		Type:    catalogv1.AttributeType_ATTRIBUTE_TYPE_TEXT,
		Enabled: true,
	})
	require.NoError(t, err)
	created := createResp.GetAttribute()

	_, err = testAttributeClient.UpdateAttribute(ctx, &catalogv1.UpdateAttributeRequest{
		Id:      created.GetId(),
		Name:    "Updated Name",
		Enabled: true,
		Version: 999,
	})
	require.Error(t, err)

	statusErr, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Aborted, statusErr.Code())
}

func int32Ptr(v int32) *int32 {
	return &v
}
