// Package migration owns Nomen's product overlay schema.
//
// Its own schema (`nomen_product`), its own version table (`nomen_product.migrations`) and
// its own sequence starting at one — deliberately separate from the compatibility
// schema next door, which belongs to `backend/v3` and is numbered 001–018 by
// upstream. Sharing their series would mean our 019 colliding with their next
// one on any future sync, and the collision would surface as a migration that
// silently does not run.
//
// Ours run immediately after theirs (see the pool's Migrate), because our
// tables carry foreign keys into compatibility-owned users and a foreign key to a table
// that does not exist yet fails at apply time rather than at review time.
package migration

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
)

var migrations []*migrate.Migration

// Migrate brings the `nomen_product` schema up to date, acquiring one connection
// from the pool for the duration.
//
// tern records the applied version in `nomen_product.migrations`, so calling this on
// every start is a cheap no-op once the schema is current — and is what keeps
// adding a migration an ordinary act rather than an operational event.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	return MigrateConn(ctx, conn.Conn())
}

// MigrateConn is Migrate against a connection the caller already holds.
func MigrateConn(ctx context.Context, conn *pgx.Conn) error {
	// The schema has to exist before the migrator can create its version table
	// inside it — the same ordering constraint v3 has, for the same reason.
	if _, err := conn.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS nomen_product"); err != nil {
		return err
	}
	migrator, err := migrate.NewMigrator(ctx, conn, "nomen_product.migrations")
	if err != nil {
		return err
	}
	migrator.Migrations = migrations
	return migrator.Migrate(ctx)
}

// registerSQLMigration is called from each migration's init.
//
// One Go file per migration, named <sequence>_<name>.go, with its SQL embedded
// from <sequence>_<name>/{up,down}.sql. Same convention as v3, so anybody who
// has read one tree can read the other.
func registerSQLMigration(sequence int32, up, down string) {
	migrations = append(migrations, &migrate.Migration{
		Sequence: sequence,
		UpSQL:    up,
		DownSQL:  down,
	})
}
