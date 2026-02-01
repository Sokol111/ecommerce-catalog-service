package product

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	productmocks "github.com/Sokol111/ecommerce-catalog-service/internal/domain/product/mocks"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

func ptr[T any](v T) *T {
	return &v
}

func createTestProductForQuery(id string) *product.Product {
	return product.Reconstruct(
		id,
		1,
		"Test Product",
		ptr("Test description"),
		99.99,
		10,
		ptr("image-123"),
		ptr("category-123"),
		true,
		nil,
		time.Now().UTC(),
		time.Now().UTC(),
	)
}

func TestGetProductByIDHandler_Handle_Success(t *testing.T) {
	repo := productmocks.NewMockRepository(t)
	handler := NewGetProductByIDHandler(repo)

	ctx := context.Background()
	productID := "product-123"
	expectedProduct := createTestProductForQuery(productID)

	repo.EXPECT().
		FindByID(mock.Anything, productID).
		Return(expectedProduct, nil)

	result, err := handler.Handle(ctx, GetProductByIDQuery{ID: productID})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expectedProduct.ID, result.ID)
	assert.Equal(t, expectedProduct.Name, result.Name)
}

func TestGetProductByIDHandler_Handle_NotFound(t *testing.T) {
	repo := productmocks.NewMockRepository(t)
	handler := NewGetProductByIDHandler(repo)

	ctx := context.Background()
	productID := "non-existent-id"

	repo.EXPECT().
		FindByID(mock.Anything, productID).
		Return(nil, persistence.ErrEntityNotFound)

	result, err := handler.Handle(ctx, GetProductByIDQuery{ID: productID})

	require.Error(t, err)
	assert.ErrorIs(t, err, persistence.ErrEntityNotFound)
	assert.Nil(t, result)
}

func TestGetProductByIDHandler_Handle_RepositoryError(t *testing.T) {
	repo := productmocks.NewMockRepository(t)
	handler := NewGetProductByIDHandler(repo)

	ctx := context.Background()
	productID := "product-123"

	repo.EXPECT().
		FindByID(mock.Anything, productID).
		Return(nil, errors.New("database error"))

	result, err := handler.Handle(ctx, GetProductByIDQuery{ID: productID})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get product")
	assert.Nil(t, result)
}

func TestGetListProductsHandler_Handle_Success(t *testing.T) {
	repo := productmocks.NewMockRepository(t)
	handler := NewGetListProductsHandler(repo)

	ctx := context.Background()
	products := []*product.Product{
		createTestProductForQuery("product-1"),
		createTestProductForQuery("product-2"),
		createTestProductForQuery("product-3"),
	}

	query := GetListProductsQuery{
		Page: 1,
		Size: 10,
	}

	repo.EXPECT().
		FindList(mock.Anything, mock.MatchedBy(func(q product.ListQuery) bool {
			return q.Page == 1 && q.Size == 10
		})).
		Return(&mongo.PageResult[product.Product]{
			Items: products,
			Page:  1,
			Size:  10,
			Total: 3,
		}, nil)

	result, err := handler.Handle(ctx, query)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Items, 3)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.Size)
	assert.Equal(t, int64(3), result.Total)
}

func TestGetListProductsHandler_Handle_WithFilters(t *testing.T) {
	repo := productmocks.NewMockRepository(t)
	handler := NewGetListProductsHandler(repo)

	ctx := context.Background()
	enabled := true
	categoryID := "category-123"

	query := GetListProductsQuery{
		Page:       2,
		Size:       5,
		Enabled:    &enabled,
		CategoryID: &categoryID,
		Sort:       "name",
		Order:      "asc",
	}

	repo.EXPECT().
		FindList(mock.Anything, mock.MatchedBy(func(q product.ListQuery) bool {
			return q.Page == 2 &&
				q.Size == 5 &&
				q.Enabled != nil && *q.Enabled == true &&
				q.CategoryID != nil && *q.CategoryID == categoryID &&
				q.Sort == "name" &&
				q.Order == "asc"
		})).
		Return(&mongo.PageResult[product.Product]{
			Items: []*product.Product{},
			Page:  2,
			Size:  5,
			Total: 0,
		}, nil)

	result, err := handler.Handle(ctx, query)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Items)
}

func TestGetListProductsHandler_Handle_RepositoryError(t *testing.T) {
	repo := productmocks.NewMockRepository(t)
	handler := NewGetListProductsHandler(repo)

	ctx := context.Background()
	query := GetListProductsQuery{
		Page: 1,
		Size: 10,
	}

	repo.EXPECT().
		FindList(mock.Anything, mock.Anything).
		Return(nil, errors.New("database error"))

	result, err := handler.Handle(ctx, query)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get products list")
	assert.Nil(t, result)
}

func TestGetListProductsHandler_Handle_EmptyResult(t *testing.T) {
	repo := productmocks.NewMockRepository(t)
	handler := NewGetListProductsHandler(repo)

	ctx := context.Background()
	query := GetListProductsQuery{
		Page: 1,
		Size: 10,
	}

	repo.EXPECT().
		FindList(mock.Anything, mock.Anything).
		Return(&mongo.PageResult[product.Product]{
			Items: []*product.Product{},
			Page:  1,
			Size:  10,
			Total: 0,
		}, nil)

	result, err := handler.Handle(ctx, query)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Items)
	assert.Equal(t, int64(0), result.Total)
}
