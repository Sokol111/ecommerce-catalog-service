//go:build integration

package mongo

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	mongooptions "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/Sokol111/ecommerce-catalog-service/internal/application/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/category"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/product"
	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/mongo"
	"github.com/Sokol111/ecommerce-commons/pkg/testutil/container"
)

var (
	testMongoContainer *container.MongoDBContainer
	testDatabase       *mongo.Database

	// Repositories for tests
	testAttributeRepo attribute.Repository
	testCategoryRepo  category.Repository
	testProductRepo   product.Repository
)

const testDBName = "catalog_test"

func TestMain(m *testing.M) {
	testMongoContainer = container.StartDefaultMongoDBContainer()
	defer testMongoContainer.Terminate()
	testDatabase = testMongoContainer.Database(testDBName)

	attributeGenericRepo, err := commonsmongo.NewGenericRepository(
		commonsmongo.NewStaticCollectionProvider(testDatabase.Collection("attribute")),
		NewAttributeMapper(),
	)
	if err != nil {
		log.Fatalf("failed to create attribute generic repository: %v", err)
	}
	// Create repositories with mappers
	testAttributeRepo, err = NewAttributeRepository(attributeGenericRepo)
	if err != nil {
		log.Fatalf("failed to create attribute repository: %v", err)
	}

	categoryGenericRepo, err := commonsmongo.NewGenericRepository(
		commonsmongo.NewStaticCollectionProvider(testDatabase.Collection("category")),
		NewCategoryMapper(),
	)
	if err != nil {
		log.Fatalf("failed to create category generic repository: %v", err)
	}

	testCategoryRepo, err = NewCategoryRepository(categoryGenericRepo)
	if err != nil {
		log.Fatalf("failed to create category repository: %v", err)
	}

	productGenericRepo, err := commonsmongo.NewGenericRepository(
		commonsmongo.NewStaticCollectionProvider(testDatabase.Collection("product")),
		NewProductMapper(),
	)
	if err != nil {
		log.Fatalf("failed to create product generic repository: %v", err)
	}

	testProductRepo, err = NewProductRepository(productGenericRepo)
	if err != nil {
		log.Fatalf("failed to create product repository: %v", err)
	}

	// Create indexes
	if err := createIndexes(context.Background()); err != nil {
		log.Fatalf("failed to create indexes: %v", err)
	}

	// Run tests
	code := m.Run()

	os.Exit(code)
}

func createIndexes(ctx context.Context) error {
	// Attribute unique slug index
	_, err := testDatabase.Collection("attribute").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]int{"slug": 1},
		Options: mongooptions.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	return nil
}

func cleanupCollection(t *testing.T, collectionName string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := testDatabase.Collection(collectionName).DeleteMany(ctx, map[string]any{})
	if err != nil {
		t.Fatalf("failed to cleanup collection %s: %v", collectionName, err)
	}
}
