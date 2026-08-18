package migration

import (
	_ "embed"
)

var (
	//go:embed 002_flows/up.sql
	up002Flows string
	//go:embed 002_flows/down.sql
	down002Flows string
)

func init() {
	registerSQLMigration(2, up002Flows, down002Flows)
}
