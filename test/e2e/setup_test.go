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

	catalogv1 "github.com/Sokol111/ecommerce-catalog-service-api/gen/go/catalog/v1"
	"github.com/Sokol111/ecommerce-commons/pkg/core/health"
	fx_commons "github.com/Sokol111/ecommerce-commons/pkg/fxconfig"
	"github.com/Sokol111/ecommerce-commons/pkg/http/grpc/client"
	"github.com/Sokol111/ecommerce-commons/pkg/testutil/container"

	fx_application "github.com/Sokol111/ecommerce-catalog-service/internal/application/fxconfig"
	fx_infrastructure "github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/fxconfig"
)

var (
	testAttributeClient catalogv1.AttributeServiceClient
	testProductClient   catalogv1.ProductServiceClient
	testCategoryClient  catalogv1.CategoryServiceClient
)

const testServerPort = 18080

func TestMain(m *testing.M) {
	mongoContainer := container.StartDefaultMongoDBContainer()
	defer mongoContainer.Terminate()
	redpandaContainer := container.StartDefaultRedpandaContainer()
	defer redpandaContainer.Terminate()
	testApp := startApp(mongoContainer, redpandaContainer)
	defer testApp.RequireStop()

	code := m.Run()

	os.Exit(code)
}

func startApp(mongoContainer *container.MongoDBContainer, redpandaContainer *container.RedpandaContainer) *fxtest.App {
	os.Setenv("APP_SERVICE_NAME", "ecommerce-catalog-service")
	os.Setenv("APP_ENV", "e2e")
	os.Setenv("APP_SERVICE_VERSION", "1.0.0")
	os.Setenv("MONGO__CONNECTION_STRING", mongoContainer.ConnectionString)
	os.Setenv("MONGO__DATABASE_NAME", "e2e_test")
	os.Setenv("MONGO__MIGRATIONS_PATH", "../../db/migrations")
	os.Setenv("SERVER__PORT", fmt.Sprintf("%d", testServerPort))
	os.Setenv("KAFKA__BROKER", redpandaContainer.KafkaBroker)

	var ready health.ReadinessWaiter

	app := fxtest.New(
		&testing.T{},

		fx.Populate(&ready),

		fx_commons.NewCommonsModule(),
		fx_application.NewAppModule(),
		fx_infrastructure.NewInfrastructureModule(),
	).RequireStart()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// Wait for all components to be ready
	defer cancel()
	if err := ready.WaitReady(ctx); err != nil {
		log.Fatalf("app not ready: %v", err)
	}

	grpcConn, err := client.NewGrpcConn(client.Config{
		Address: fmt.Sprintf("localhost:%d", testServerPort),
		Timeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatalf("failed to create gRPC connection: %v", err)
	}
	testAttributeClient = catalogv1.NewAttributeServiceClient(grpcConn)
	testProductClient = catalogv1.NewProductServiceClient(grpcConn)
	testCategoryClient = catalogv1.NewCategoryServiceClient(grpcConn)
	return app
}
