package migration

import (
	_ "embed"
)

var (
	//go:embed 001_seats/up.sql
	up001Seats string
	//go:embed 001_seats/down.sql
	down001Seats string
)

func init() {
	registerSQLMigration(1, up001Seats, down001Seats)
}
