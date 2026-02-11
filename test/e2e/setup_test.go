//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/httpapi"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application"
	internalhttp "github.com/Sokol111/ecommerce-catalog-service/internal/http"
	"github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/persistence/mongo"
	commons_core "github.com/Sokol111/ecommerce-commons/pkg/core"
	"github.com/Sokol111/ecommerce-commons/pkg/core/config"
	"github.com/Sokol111/ecommerce-commons/pkg/core/health"
	commons_http "github.com/Sokol111/ecommerce-commons/pkg/http"
	"github.com/Sokol111/ecommerce-commons/pkg/security/token"
	"github.com/Sokol111/ecommerce-commons/pkg/testutil/container"

	"github.com/Sokol111/ecommerce-commons/pkg/http/server"
	commons_messaging "github.com/Sokol111/ecommerce-commons/pkg/messaging"
	kafka_config "github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/config"
	commons_observability "github.com/Sokol111/ecommerce-commons/pkg/observability"
	commons_persistence "github.com/Sokol111/ecommerce-commons/pkg/persistence"
	commons_mongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	commons_security "github.com/Sokol111/ecommerce-commons/pkg/security"
)

var (
	testApp                     *fxtest.App
	testServerURL               string
	testClient                  *httpapi.Client
	testMongoContainer          *container.MongoDBContainer
	testSchemaRegistryContainer *container.SchemaRegistryContainer
	testReadinessWaiter         health.ReadinessWaiter
)

const testServerPort = 18080

func TestMain(m *testing.M) {
	ctx := context.Background()

	startContainers(ctx)
	startApp(ctx)
	createTestClient()

	code := m.Run()

	stopApp()
	stopContainers()

	os.Exit(code)
}

func startContainers(ctx context.Context) {
	var err error

	// Start MongoDB container
	testMongoContainer, err = container.StartMongoDBContainer(ctx, container.WithReplicaSet("rs0"))
	if err != nil {
		log.Fatalf("failed to start mongodb container: %v", err)
	}

	// Start Schema Registry container (Redpanda with embedded Kafka)
	testSchemaRegistryContainer, err = container.StartSchemaRegistryContainer(ctx)
	if err != nil {
		log.Fatalf("failed to start schema registry container: %v", err)
	}
}

func stopContainers() {
	ctx := context.Background()
	if err := testMongoContainer.Terminate(ctx); err != nil {
		log.Printf("failed to terminate mongodb: %v", err)
	}
	if err := testSchemaRegistryContainer.Terminate(ctx); err != nil {
		log.Printf("failed to terminate schema registry: %v", err)
	}
}

func startApp(ctx context.Context) {
	kafkaBroker, err := testSchemaRegistryContainer.KafkaBroker(ctx)
	if err != nil {
		log.Fatalf("failed to get kafka broker: %v", err)
	}

	testApp = fxtest.New(
		&testing.T{},

		// Extract ReadinessWaiter from DI
		fx.Populate(&testReadinessWaiter),

		// Commons modules with test configs
		commons_core.NewCoreModule(
			commons_core.WithAppConfig(
				config.AppConfig{
					ServiceName:    "ecommerce-catalog-service",
					Environment:    "test",
					ServiceVersion: "1.0.0",
				},
			),
			commons_core.WithoutConfigFile(),
			commons_core.WithoutEnvFile(),
		),
		commons_persistence.NewPersistenceModule(
			commons_persistence.WithMongoConfig(
				commons_mongo.Config{
					ConnectionString: testMongoContainer.ConnectionString,
					Database:         "catalog_e2e_test",
				},
			),
		),
		commons_http.NewHTTPModule(
			commons_http.WithServerConfig(
				server.Config{
					Port: testServerPort,
				},
			),
		),
		commons_observability.NewObservabilityModule(
			commons_observability.WithoutMetrics(),
			commons_observability.WithoutTracing(),
		),
		commons_messaging.NewMessagingModule(
			commons_messaging.WithKafkaConfig(kafka_config.Config{
				Brokers: kafkaBroker,
				SchemaRegistry: kafka_config.SchemaRegistryConfig{
					URL: testSchemaRegistryContainer.URL,
				},
			}),
		),
		commons_security.NewSecurityModule(
			commons_security.WithTestValidator(),
		),

		// Application modules
		mongo.Module(),
		application.Module(),
		internalhttp.Module(),
	)

	testApp.RequireStart()

	// Wait for all components to be ready
	readyCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = testReadinessWaiter.WaitReady(readyCtx); err != nil {
		log.Fatalf("app not ready: %v", err)
	}

	testServerURL = fmt.Sprintf("http://localhost:%d", testServerPort)
}

func createTestClient() {
	var err error
	testClient, err = httpapi.NewClient(testServerURL, &testSecuritySource{
		token: token.GenerateAdminTestToken(),
	})
	if err != nil {
		log.Fatalf("failed to create test client: %v", err)
	}
}

func stopApp() {
	testApp.RequireStop()
}

// testSecuritySource provides test tokens for the HTTP client.
type testSecuritySource struct {
	token string
}

func (s *testSecuritySource) BearerAuth(context.Context, httpapi.OperationName) (httpapi.BearerAuth, error) {
	return httpapi.BearerAuth{Token: s.token}, nil
}

func cleanupDatabase(t *testing.T) {
	t.Helper()
	// Implement database cleanup between tests if needed
	// Can use testClient to delete all entities or direct DB access
}
