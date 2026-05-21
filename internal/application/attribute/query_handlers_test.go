package attribute

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

// Helper function to create a test attribute
func createTestAttributeWithParams(id, name, slug string, attrType AttributeType, enabled bool) *Attribute {
	return Reconstruct(
		id,
		1,
		name,
		slug,
		attrType,
		nil, // unit
		enabled,
		[]Option{
			{Name: "Option 1", Slug: "option-1"},
			{Name: "Option 2", Slug: "option-2"},
		},
		time.Now().UTC(),
		time.Now().UTC(),
	)
}

// === GetAttributeByIDHandler Tests ===

func TestGetAttributeByIDHandler_Handle_Success(t *testing.T) {
	repo := NewMockRepository(t)
	handler := NewGetAttributeByIDHandler(repo)

	ctx := context.Background()
	expectedAttr := createTestAttributeWithParams("attr-123", "Color", "color", AttributeTypeSingle, true)

	repo.EXPECT().
		FindByID(mock.Anything, "attr-123").
		Return(expectedAttr, nil)

	result, err := handler.Handle(ctx, GetAttributeByIDQuery{ID: "attr-123"})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expectedAttr.ID, result.ID)
	assert.Equal(t, expectedAttr.Name, result.Name)
	assert.Equal(t, expectedAttr.Slug, result.Slug)
	assert.True(t, result.Enabled)
}

func TestGetAttributeByIDHandler_Handle_NotFound(t *testing.T) {
	repo := NewMockRepository(t)
	handler := NewGetAttributeByIDHandler(repo)

	ctx := context.Background()

	repo.EXPECT().
		FindByID(mock.Anything, "non-existent-id").
		Return(nil, commonsmongo.ErrEntityNotFound)

	result, err := handler.Handle(ctx, GetAttributeByIDQuery{ID: "non-existent-id"})

	require.Error(t, err)
	assert.ErrorIs(t, err, commonsmongo.ErrEntityNotFound)
	assert.Nil(t, result)
}

func TestGetAttributeByIDHandler_Handle_RepositoryError(t *testing.T) {
	repo := NewMockRepository(t)
	handler := NewGetAttributeByIDHandler(repo)

	ctx := context.Background()

	repo.EXPECT().
		FindByID(mock.Anything, "attr-123").
		Return(nil, errors.New("database connection error"))

	result, err := handler.Handle(ctx, GetAttributeByIDQuery{ID: "attr-123"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get attribute")
	assert.Nil(t, result)
}

// === GetAttributeListHandler Tests ===

func TestGetAttributeListHandler_Handle_Success(t *testing.T) {
	repo := NewMockRepository(t)
	handler := NewGetAttributeListHandler(repo)

	ctx := context.Background()
	expectedAttributes := []*Attribute{
		createTestAttributeWithParams("attr-1", "Color", "color", AttributeTypeSingle, true),
		createTestAttributeWithParams("attr-2", "Size", "size", AttributeTypeSingle, true),
		createTestAttributeWithParams("attr-3", "Material", "material", AttributeTypeMultiple, false),
	}

	repo.EXPECT().
		FindList(mock.Anything, mock.MatchedBy(func(q ListQuery) bool {
			return q.Page == 1 && q.Size == 10
		})).
		Return(&commonsmongo.PageResult[Attribute]{
			Items: expectedAttributes,
			Page:  1,
			Size:  10,
			Total: 3,
		}, nil)

	result, err := handler.Handle(ctx, GetAttributeListQuery{
		Page: 1,
		Size: 10,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Items, 3)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.Size)
	assert.Equal(t, int64(3), result.Total)
}

func TestGetAttributeListHandler_Handle_WithEnabledFilter(t *testing.T) {
	repo := NewMockRepository(t)
	handler := NewGetAttributeListHandler(repo)

	ctx := context.Background()
	enabled := true
	expectedAttributes := []*Attribute{
		createTestAttributeWithParams("attr-1", "Color", "color", AttributeTypeSingle, true),
		createTestAttributeWithParams("attr-2", "Size", "size", AttributeTypeSingle, true),
	}

	repo.EXPECT().
		FindList(mock.Anything, mock.MatchedBy(func(q ListQuery) bool {
			return q.Enabled != nil && *q.Enabled == true
		})).
		Return(&commonsmongo.PageResult[Attribute]{
			Items: expectedAttributes,
			Page:  1,
			Size:  10,
			Total: 2,
		}, nil)

	result, err := handler.Handle(ctx, GetAttributeListQuery{
		Page:    1,
		Size:    10,
		Enabled: &enabled,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(2), result.Total)
}

func TestGetAttributeListHandler_Handle_WithTypeFilter(t *testing.T) {
	repo := NewMockRepository(t)
	handler := NewGetAttributeListHandler(repo)

	ctx := context.Background()
	attrType := string(AttributeTypeSingle)
	expectedAttributes := []*Attribute{
		createTestAttributeWithParams("attr-1", "Color", "color", AttributeTypeSingle, true),
		createTestAttributeWithParams("attr-2", "Size", "size", AttributeTypeSingle, true),
	}

	repo.EXPECT().
		FindList(mock.Anything, mock.MatchedBy(func(q ListQuery) bool {
			return q.Type != nil && *q.Type == "single"
		})).
		Return(&commonsmongo.PageResult[Attribute]{
			Items: expectedAttributes,
			Page:  1,
			Size:  10,
			Total: 2,
		}, nil)

	result, err := handler.Handle(ctx, GetAttributeListQuery{
		Page: 1,
		Size: 10,
		Type: &attrType,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Items, 2)
}

func TestGetAttributeListHandler_Handle_WithSorting(t *testing.T) {
	repo := NewMockRepository(t)
	handler := NewGetAttributeListHandler(repo)

	ctx := context.Background()
	expectedAttributes := []*Attribute{
		createTestAttributeWithParams("attr-1", "Color", "color", AttributeTypeSingle, true),
		createTestAttributeWithParams("attr-3", "Material", "material", AttributeTypeMultiple, true),
		createTestAttributeWithParams("attr-2", "Size", "size", AttributeTypeSingle, true),
	}

	repo.EXPECT().
		FindList(mock.Anything, mock.MatchedBy(func(q ListQuery) bool {
			return q.Sort == "name" && q.Order == "asc"
		})).
		Return(&commonsmongo.PageResult[Attribute]{
			Items: expectedAttributes,
			Page:  1,
			Size:  10,
			Total: 3,
		}, nil)

	result, err := handler.Handle(ctx, GetAttributeListQuery{
		Page:  1,
		Size:  10,
		Sort:  "name",
		Order: "asc",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Items, 3)
}

func TestGetAttributeListHandler_Handle_EmptyResult(t *testing.T) {
	repo := NewMockRepository(t)
	handler := NewGetAttributeListHandler(repo)

	ctx := context.Background()

	repo.EXPECT().
		FindList(mock.Anything, mock.Anything).
		Return(&commonsmongo.PageResult[Attribute]{
			Items: []*Attribute{},
			Page:  1,
			Size:  10,
			Total: 0,
		}, nil)

	result, err := handler.Handle(ctx, GetAttributeListQuery{
		Page: 1,
		Size: 10,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Items)
	assert.Equal(t, int64(0), result.Total)
}

func TestGetAttributeListHandler_Handle_Pagination(t *testing.T) {
	repo := NewMockRepository(t)
	handler := NewGetAttributeListHandler(repo)

	ctx := context.Background()
	expectedAttributes := []*Attribute{
		createTestAttributeWithParams("attr-5", "Weight", "weight", AttributeTypeSingle, true),
		createTestAttributeWithParams("attr-6", "Height", "height", AttributeTypeSingle, true),
	}

	repo.EXPECT().
		FindList(mock.Anything, mock.MatchedBy(func(q ListQuery) bool {
			return q.Page == 3 && q.Size == 2
		})).
		Return(&commonsmongo.PageResult[Attribute]{
			Items: expectedAttributes,
			Page:  3,
			Size:  2,
			Total: 20,
		}, nil)

	result, err := handler.Handle(ctx, GetAttributeListQuery{
		Page: 3,
		Size: 2,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 3, result.Page)
	assert.Equal(t, 2, result.Size)
	assert.Equal(t, int64(20), result.Total)
}

func TestGetAttributeListHandler_Handle_CombinedFilters(t *testing.T) {
	repo := NewMockRepository(t)
	handler := NewGetAttributeListHandler(repo)

	ctx := context.Background()
	enabled := true
	attrType := string(AttributeTypeSingle)
	expectedAttributes := []*Attribute{
		createTestAttributeWithParams("attr-1", "Color", "color", AttributeTypeSingle, true),
	}

	repo.EXPECT().
		FindList(mock.Anything, mock.MatchedBy(func(q ListQuery) bool {
			return q.Enabled != nil && *q.Enabled == true &&
				q.Type != nil && *q.Type == "single" &&
				q.Sort == "name" && q.Order == "desc"
		})).
		Return(&commonsmongo.PageResult[Attribute]{
			Items: expectedAttributes,
			Page:  1,
			Size:  10,
			Total: 1,
		}, nil)

	result, err := handler.Handle(ctx, GetAttributeListQuery{
		Page:    1,
		Size:    10,
		Enabled: &enabled,
		Type:    &attrType,
		Sort:    "name",
		Order:   "desc",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Items, 1)
}

func TestGetAttributeListHandler_Handle_RepositoryError(t *testing.T) {
	repo := NewMockRepository(t)
	handler := NewGetAttributeListHandler(repo)

	ctx := context.Background()

	repo.EXPECT().
		FindList(mock.Anything, mock.Anything).
		Return(nil, errors.New("database error"))

	result, err := handler.Handle(ctx, GetAttributeListQuery{
		Page: 1,
		Size: 10,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get attributes list")
	assert.Nil(t, result)
}
