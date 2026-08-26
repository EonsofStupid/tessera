//go:build integration

package events_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shippinAI/nomen/backend/v3/storage/database"
	"github.com/shippinAI/nomen/backend/v3/storage/database/dialect/postgres"
	"github.com/shippinAI/nomen/internal/integration"
	"github.com/shippinAI/nomen/pkg/grpc/admin"
	"github.com/shippinAI/nomen/pkg/grpc/authorization/v2"
	v2beta "github.com/shippinAI/nomen/pkg/grpc/instance/v2beta"
	mgmt "github.com/shippinAI/nomen/pkg/grpc/management"
	v2beta_org "github.com/shippinAI/nomen/pkg/grpc/org/v2beta"
	v2beta_project "github.com/shippinAI/nomen/pkg/grpc/project/v2beta"
	"github.com/shippinAI/nomen/pkg/grpc/session/v2"
	"github.com/shippinAI/nomen/pkg/grpc/system"
	"github.com/shippinAI/nomen/pkg/grpc/user/v2"
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

var ConnString = fmt.Sprintf(
	"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
	getEnv("NOMEN_DATABASE_POSTGRES_HOST", "localhost"),
	getEnv("NOMEN_DATABASE_POSTGRES_PORT", "5433"),
	getEnv("NOMEN_DATABASE_POSTGRES_USER", "nomen"),
	getEnv("NOMEN_DATABASE_POSTGRES_PASSWORD", "nomen"),
	getEnv("NOMEN_DATABASE_POSTGRES_DATABASE", "nomen"),
	getEnv("NOMEN_DATABASE_POSTGRES_SSL_MODE", "disable"),
)

var (
	dbPool              *pgxpool.Pool
	CTX                 context.Context
	IAMCTX              context.Context
	Instance            *integration.Instance
	SystemClient        system.SystemServiceClient
	OrgClient           v2beta_org.OrganizationServiceClient
	ProjectClient       v2beta_project.ProjectServiceClient
	SessionClient       session.SessionServiceClient
	UserClient          user.UserServiceClient
	AdminClient         admin.AdminServiceClient
	MgmtClient          mgmt.ManagementServiceClient
	AuthorizationClient authorization.AuthorizationServiceClient
)

var pool database.Pool

func TestMain(m *testing.M) {
	os.Exit(func() int {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		CTX = integration.WithSystemAuthorization(ctx)
		Instance = integration.NewInstance(CTX)

		IAMCTX = Instance.WithAuthorizationToken(ctx, integration.UserTypeIAMOwner)
		SystemClient = integration.SystemClient()
		OrgClient = Instance.Client.OrgV2beta
		ProjectClient = Instance.Client.Projectv2Beta
		SessionClient = Instance.Client.SessionV2
		UserClient = Instance.Client.UserV2
		AdminClient = Instance.Client.Admin
		MgmtClient = Instance.Client.Mgmt
		AuthorizationClient = Instance.Client.AuthorizationV2

		defer func() {
			_, err := Instance.Client.InstanceV2Beta.DeleteInstance(CTX, &v2beta.DeleteInstanceRequest{
				InstanceId: Instance.Instance.Id,
			})
			if err != nil {
				log.Printf("Failed to delete instance on cleanup: %v\n", err)
			}
		}()

		var err error
		dbConfig, err := pgxpool.ParseConfig(ConnString)
		if err != nil {
			panic(err)
		}
		dbConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			orgState, err := conn.LoadType(ctx, "nomen.organization_state")
			if err != nil {
				return err
			}
			conn.TypeMap().RegisterType(orgState)
			return nil
		}

		dbPool, err = pgxpool.NewWithConfig(context.Background(), dbConfig)
		if err != nil {
			panic(err)
		}

		pool = postgres.PGxPool(dbPool)

		return m.Run()
	}())
}
