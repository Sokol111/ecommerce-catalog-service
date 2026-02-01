//go:build integration

package mongo

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

var (
	testClient   *mongo.Client
	testDatabase *mongo.Database
	testMongo    commonsmongo.Mongo

	// Repositories for tests
	testAttributeRepo attribute.Repository
	testCategoryRepo  category.Repository
	testProductRepo   product.Repository
)

const testDBName = "catalog_test"

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start MongoDB container
	mongoContainer, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		log.Fatalf("failed to start mongodb container: %v", err)
	}

	// Get connection string
	connectionString, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	// Connect to MongoDB
	clientOpts := options.Client().ApplyURI(connectionString)
	testClient, err = mongo.Connect(clientOpts)
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}

	// Ping to verify connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := testClient.Ping(ctx, nil); err != nil {
		log.Fatalf("failed to ping mongodb: %v", err)
	}

	testDatabase = testClient.Database(testDBName)
	testMongo = &testMongoWrapper{db: testDatabase}

	// Create repositories with mappers
	testAttributeRepo, err = newAttributeRepository(testMongo, newAttributeMapper())
	if err != nil {
		log.Fatalf("failed to create attribute repository: %v", err)
	}

	testCategoryRepo, err = newCategoryRepository(testMongo, newCategoryMapper())
	if err != nil {
		log.Fatalf("failed to create category repository: %v", err)
	}

	testProductRepo, err = newProductRepository(testMongo, newProductMapper())
	if err != nil {
		log.Fatalf("failed to create product repository: %v", err)
	}

	// Create indexes
	if err := createIndexes(context.Background()); err != nil {
		log.Fatalf("failed to create indexes: %v", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if err := testClient.Disconnect(context.Background()); err != nil {
		log.Printf("failed to disconnect from mongodb: %v", err)
	}
	if err := testcontainers.TerminateContainer(mongoContainer); err != nil {
		log.Printf("failed to terminate mongodb container: %v", err)
	}

	os.Exit(code)
}

// testMongoWrapper implements commonsmongo.Mongo interface
type testMongoWrapper struct {
	db *mongo.Database
}

func (m *testMongoWrapper) GetCollection(name string) *mongo.Collection {
	return m.db.Collection(name)
}

func createIndexes(ctx context.Context) error {
	// Attribute unique slug index
	_, err := testDatabase.Collection("attribute").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]int{"slug": 1},
		Options: options.Index().SetUnique(true),
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

func ptrI[T any](v T) *T {
	return &v
}
