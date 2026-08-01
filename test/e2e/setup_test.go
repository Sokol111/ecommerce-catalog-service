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

	fx_inbound "github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/inbound/fxconfig"
	fx_outbound "github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/outbound/fxconfig"

	catalogv1 "github.com/Sokol111/ecommerce-catalog-service-api/gen/go/catalog/v1"
	"github.com/Sokol111/ecommerce-commons/pkg/core/health"
	fx_commons "github.com/Sokol111/ecommerce-commons/pkg/fxconfig"
	"github.com/Sokol111/ecommerce-commons/pkg/http/grpc/client"
	"github.com/Sokol111/ecommerce-commons/pkg/tenant"
	"github.com/Sokol111/ecommerce-commons/pkg/testutil/container"

	fx_application "github.com/Sokol111/ecommerce-catalog-service/internal/application/fxconfig"
)

var (
	testAttributeClient catalogv1.AttributeServiceClient
	testProductClient   catalogv1.ProductServiceClient
	testCategoryClient  catalogv1.CategoryServiceClient
)

const testServerPort = 18080

type defaultSlugsProvider struct{}

func (defaultSlugsProvider) GetSlugs(ctx context.Context) ([]string, error) {
	return []string{"default"}, nil
}

func TestMain(m *testing.M) {
	mongoContainer := container.StartDefaultMongoDBContainer()
	defer mongoContainer.Terminate()
	redpandaContainer := container.StartDefaultRedpandaContainer()
	defer redpandaContainer.Terminate()
	testApp := startApp(mongoContainer, redpandaContainer)
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := testApp.Stop(stopCtx); err != nil {
			log.Printf("failed to stop app: %v", err)
		}
	}()

	code := m.Run()

	os.Exit(code)
}

func startApp(mongoContainer *container.MongoDBContainer, redpandaContainer *container.RedpandaContainer) *fx.App {
	os.Setenv("APP_SERVICE_NAME", "ecommerce-catalog-service")
	os.Setenv("APP_ENV", "e2e")
	os.Setenv("APP_SERVICE_VERSION", "1.0.0")
	os.Setenv("CONFIG_FILE", "../../configs/config.e2e.yaml")
	os.Setenv("MONGO__CONNECTION_STRING", mongoContainer.ConnectionString)
	os.Setenv("MONGO__MIGRATIONS__PATH", "../../db/migrations")
	os.Setenv("SERVER__PORT", fmt.Sprintf("%d", testServerPort))
	os.Setenv("KAFKA__BROKERS", redpandaContainer.KafkaBroker)

	log.Println("starting fx app...")

	var ready health.ReadinessWaiter
	app := fx.New(
		fx_commons.NewCommonsModule(),
		fx_application.NewAppModule(),

		fx_outbound.NewOutboundInfrastructureModule(),
		fx_inbound.NewInboundInfrastructureModule(),
		fx.Provide(func() tenant.SlugsProvider { return defaultSlugsProvider{} }),
		fx.Populate(&ready),
	)

	startCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		log.Fatalf("failed to start app: %v", err)
	}
	log.Println("fx app started")

	waitCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := ready.WaitReady(waitCtx); err != nil {
		log.Fatalf("app not ready: %v", err)
	}
	log.Println("app is ready")

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
	log.Println("gRPC clients created")

	return app
}
