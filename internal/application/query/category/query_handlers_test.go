package category

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	categorymocks "github.com/Sokol111/ecommerce-catalog-service/internal/domain/category/mocks"
	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

// Helper function to create a test category
func createTestCategory(id, name string, enabled bool) *category.Category {
	return category.Reconstruct(
		id,
		1,
		name,
		enabled,
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

// === GetCategoryByIDHandler Tests ===

func TestGetCategoryByIDHandler_Handle_Success(t *testing.T) {
	repo := categorymocks.NewMockRepository(t)
	handler := NewGetCategoryByIDHandler(repo)

	ctx := context.Background()
	expectedCategory := createTestCategory("category-123", "Electronics", true)

	repo.EXPECT().
		FindByID(mock.Anything, "category-123").
		Return(expectedCategory, nil)

	result, err := handler.Handle(ctx, GetCategoryByIDQuery{ID: "category-123"})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expectedCategory.ID, result.ID)
	assert.Equal(t, expectedCategory.Name, result.Name)
	assert.True(t, result.Enabled)
}

func TestGetCategoryByIDHandler_Handle_NotFound(t *testing.T) {
	repo := categorymocks.NewMockRepository(t)
	handler := NewGetCategoryByIDHandler(repo)

	ctx := context.Background()

	repo.EXPECT().
		FindByID(mock.Anything, "non-existent-id").
		Return(nil, commonsmongo.ErrEntityNotFound)

	result, err := handler.Handle(ctx, GetCategoryByIDQuery{ID: "non-existent-id"})

	require.Error(t, err)
	assert.ErrorIs(t, err, commonsmongo.ErrEntityNotFound)
	assert.Nil(t, result)
}

func TestGetCategoryByIDHandler_Handle_RepositoryError(t *testing.T) {
	repo := categorymocks.NewMockRepository(t)
	handler := NewGetCategoryByIDHandler(repo)

	ctx := context.Background()

	repo.EXPECT().
		FindByID(mock.Anything, "category-123").
		Return(nil, errors.New("database connection error"))

	result, err := handler.Handle(ctx, GetCategoryByIDQuery{ID: "category-123"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get category")
	assert.Nil(t, result)
}

// === GetListCategoriesHandler Tests ===

func TestGetListCategoriesHandler_Handle_Success(t *testing.T) {
	repo := categorymocks.NewMockRepository(t)
	handler := NewGetListCategoriesHandler(repo)

	ctx := context.Background()
	expectedCategories := []*category.Category{
		createTestCategory("cat-1", "Electronics", true),
		createTestCategory("cat-2", "Clothing", true),
		createTestCategory("cat-3", "Books", false),
	}

	repo.EXPECT().
		FindList(mock.Anything, mock.MatchedBy(func(q category.ListQuery) bool {
			return q.Page == 1 && q.Size == 10
		})).
		Return(&commonsmongo.PageResult[category.Category]{
			Items: expectedCategories,
			Page:  1,
			Size:  10,
			Total: 3,
		}, nil)

	result, err := handler.Handle(ctx, GetListCategoriesQuery{
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

func TestGetListCategoriesHandler_Handle_WithEnabledFilter(t *testing.T) {
	repo := categorymocks.NewMockRepository(t)
	handler := NewGetListCategoriesHandler(repo)

	ctx := context.Background()
	enabled := true
	expectedCategories := []*category.Category{
		createTestCategory("cat-1", "Electronics", true),
		createTestCategory("cat-2", "Clothing", true),
	}

	repo.EXPECT().
		FindList(mock.Anything, mock.MatchedBy(func(q category.ListQuery) bool {
			return q.Enabled != nil && *q.Enabled == true
		})).
		Return(&commonsmongo.PageResult[category.Category]{
			Items: expectedCategories,
			Page:  1,
			Size:  10,
			Total: 2,
		}, nil)

	result, err := handler.Handle(ctx, GetListCategoriesQuery{
		Page:    1,
		Size:    10,
		Enabled: &enabled,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(2), result.Total)
}

func TestGetListCategoriesHandler_Handle_WithSorting(t *testing.T) {
	repo := categorymocks.NewMockRepository(t)
	handler := NewGetListCategoriesHandler(repo)

	ctx := context.Background()
	expectedCategories := []*category.Category{
		createTestCategory("cat-3", "Books", true),
		createTestCategory("cat-2", "Clothing", true),
		createTestCategory("cat-1", "Electronics", true),
	}

	repo.EXPECT().
		FindList(mock.Anything, mock.MatchedBy(func(q category.ListQuery) bool {
			return q.Sort == "name" && q.Order == "asc"
		})).
		Return(&commonsmongo.PageResult[category.Category]{
			Items: expectedCategories,
			Page:  1,
			Size:  10,
			Total: 3,
		}, nil)

	result, err := handler.Handle(ctx, GetListCategoriesQuery{
		Page:  1,
		Size:  10,
		Sort:  "name",
		Order: "asc",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Items, 3)
}

func TestGetListCategoriesHandler_Handle_EmptyResult(t *testing.T) {
	repo := categorymocks.NewMockRepository(t)
	handler := NewGetListCategoriesHandler(repo)

	ctx := context.Background()

	repo.EXPECT().
		FindList(mock.Anything, mock.Anything).
		Return(&commonsmongo.PageResult[category.Category]{
			Items: []*category.Category{},
			Page:  1,
			Size:  10,
			Total: 0,
		}, nil)

	result, err := handler.Handle(ctx, GetListCategoriesQuery{
		Page: 1,
		Size: 10,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Items)
	assert.Equal(t, int64(0), result.Total)
}

func TestGetListCategoriesHandler_Handle_Pagination(t *testing.T) {
	repo := categorymocks.NewMockRepository(t)
	handler := NewGetListCategoriesHandler(repo)

	ctx := context.Background()
	expectedCategories := []*category.Category{
		createTestCategory("cat-3", "Third", true),
		createTestCategory("cat-4", "Fourth", true),
	}

	repo.EXPECT().
		FindList(mock.Anything, mock.MatchedBy(func(q category.ListQuery) bool {
			return q.Page == 2 && q.Size == 2
		})).
		Return(&commonsmongo.PageResult[category.Category]{
			Items: expectedCategories,
			Page:  2,
			Size:  2,
			Total: 10,
		}, nil)

	result, err := handler.Handle(ctx, GetListCategoriesQuery{
		Page: 2,
		Size: 2,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 2, result.Page)
	assert.Equal(t, 2, result.Size)
	assert.Equal(t, int64(10), result.Total)
}

func TestGetListCategoriesHandler_Handle_RepositoryError(t *testing.T) {
	repo := categorymocks.NewMockRepository(t)
	handler := NewGetListCategoriesHandler(repo)

	ctx := context.Background()

	repo.EXPECT().
		FindList(mock.Anything, mock.Anything).
		Return(nil, errors.New("database error"))

	result, err := handler.Handle(ctx, GetListCategoriesQuery{
		Page: 1,
		Size: 10,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get categories list")
	assert.Nil(t, result)
}
