//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"go.uber.org/fx/fxtest"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/httpapi"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application"
	internalhttp "github.com/Sokol111/ecommerce-catalog-service/internal/http"
	"github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/persistence/mongo"
	"github.com/Sokol111/ecommerce-catalog-service/test/testutil"
	commons_core "github.com/Sokol111/ecommerce-commons/pkg/core"
	"github.com/Sokol111/ecommerce-commons/pkg/core/config"
	commons_http "github.com/Sokol111/ecommerce-commons/pkg/http"

	"github.com/Sokol111/ecommerce-commons/pkg/http/server"
	commons_messaging "github.com/Sokol111/ecommerce-commons/pkg/messaging"
	commons_observability "github.com/Sokol111/ecommerce-commons/pkg/observability"
	commons_persistence "github.com/Sokol111/ecommerce-commons/pkg/persistence"
	commons_mongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	commons_security "github.com/Sokol111/ecommerce-commons/pkg/security"
)

var (
	testApp            *fxtest.App
	testServerURL      string
	testClient         *httpapi.Client
	testMongoContainer *testutil.MongoDBContainer
)

const testServerPort = 18080

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. Start MongoDB container
	var err error
	testMongoContainer, err = testutil.StartMongoDBContainer(ctx, testutil.WithReplicaSet("rs0"))
	if err != nil {
		log.Fatalf("failed to start mongodb container: %v", err)
	}

	os.Setenv("KAFKA_ENABLED", "false")

	testApp = fxtest.New(
		&testing.T{},

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
		commons_messaging.NewMessagingModule(),
		commons_security.NewSecurityModule(
			commons_security.WithoutSecurity(),
		),

		// Application modules
		mongo.Module(),
		application.Module(),
		internalhttp.Module(),
	)

	testApp.RequireStart()

	testServerURL = fmt.Sprintf("http://localhost:%d", testServerPort)
	if err := waitForServer(testServerURL, 10*time.Second); err != nil {
		log.Fatalf("server did not start: %v", err)
	}

	testClient, err = httpapi.NewClient(testServerURL, &noopSecuritySource{})
	if err != nil {
		log.Fatalf("failed to create test client: %v", err)
	}

	code := m.Run()

	testApp.RequireStop()
	if err := testMongoContainer.Terminate(context.Background()); err != nil {
		log.Printf("failed to terminate mongodb: %v", err)
	}

	os.Exit(code)
}

// noopSecuritySource implements SecuritySource for the client
type noopSecuritySource struct{}

func (s *noopSecuritySource) BearerAuth(ctx context.Context, operationName httpapi.OperationName) (httpapi.BearerAuth, error) {
	return httpapi.BearerAuth{Token: "test-token"}, nil
}

func waitForServer(url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("server not ready after %v", timeout)
		default:
			resp, err := http.Get(url + "/health/ready")
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return nil
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func cleanupDatabase(t *testing.T) {
	t.Helper()
	// Implement database cleanup between tests if needed
	// Can use testClient to delete all entities or direct DB access
}
