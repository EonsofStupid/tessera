package database

import "context"

type Migrator interface {
	// Migrate executes migrations to set up the database.
	// The method can be called once per running Nomen.
	Migrate(ctx context.Context) error
}
