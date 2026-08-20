package migration

import _ "embed"

var (
	//go:embed 003_operator_events/up.sql
	up003OperatorEvents string
	//go:embed 003_operator_events/down.sql
	down003OperatorEvents string
)

func init() {
	registerSQLMigration(3, up003OperatorEvents, down003OperatorEvents)
}
