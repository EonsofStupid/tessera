package migration_test

import (
	"context"
	"log"
	"net"
	"os"
	"slices"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	nomenmigration "github.com/shippinAI/nomen/backend/v1/storage/migration"
	"github.com/shippinAI/nomen/backend/v3/storage/database"
	v3postgres "github.com/shippinAI/nomen/backend/v3/storage/database/dialect/postgres"
)

func TestIdentityTablesRecreateFromEmptyPostgres(t *testing.T) {
	ctx := context.Background()
	tmp, err := os.MkdirTemp("", "nomen-recreate-*")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tmp) })

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := uint32(listener.Addr().(*net.TCPAddr).Port)
	require.NoError(t, listener.Close())

	cfg := embeddedpostgres.DefaultConfig().Version(embeddedpostgres.V16).
		Port(port).RuntimePath(tmp).StartTimeout(90 * time.Second)
	pg := embeddedpostgres.NewDatabase(cfg)
	if err := pg.Start(); err != nil {
		t.Skipf("embedded postgres unavailable (%v) — run: bash dev/preflight.sh --fetch-embedded", err)
	}
	t.Cleanup(func() { _ = pg.Stop() })

	url := cfg.GetConnectionURL() + "?sslmode=disable"
	first := applyIdentityMigrations(t, ctx, url)
	require.NotEmpty(t, first, "first migrate produced no identity tables")

	pool, err := pgxpool.New(ctx, url)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `DROP SCHEMA IF EXISTS nomen CASCADE; DROP SCHEMA IF EXISTS nomen_product CASCADE`)
	require.NoError(t, err)
	pool.Close()

	second := applyIdentityMigrations(t, ctx, url)
	require.Equal(t, first, second, "dropping schemas and migrating again must recreate the same tables")
}

func applyIdentityMigrations(t *testing.T, ctx context.Context, url string) []string {
	t.Helper()
	connector, err := v3postgres.DecodeConfig(url)
	require.NoError(t, err)
	client, err := connector.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, client.(database.PoolTest).MigrateTest(ctx))

	raw, err := pgxpool.New(ctx, url)
	require.NoError(t, err)
	defer raw.Close()
	require.NoError(t, nomenmigration.Migrate(ctx, raw))

	rows, err := raw.Query(ctx, `
SELECT schemaname || '.' || tablename
FROM pg_catalog.pg_tables
WHERE schemaname IN ('nomen', 'nomen_product')
ORDER BY 1`)
	require.NoError(t, err)
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		names = append(names, name)
	}
	require.NoError(t, rows.Err())
	if !slices.Contains(names, "nomen.users") || !slices.Contains(names, "nomen_product.nomen_owner_enrollments") {
		log.Printf("tables: %v", names)
	}
	require.Contains(t, names, "nomen.instances")
	require.Contains(t, names, "nomen.users")
	require.Contains(t, names, "nomen_product.seats")
	require.Contains(t, names, "nomen_product.nomen_owner_enrollments")
	return names
}
