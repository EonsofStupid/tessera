package migration

import _ "embed"

var (
	//go:embed 004_owner_enrollment/up.sql
	up004OwnerEnrollment string
	//go:embed 004_owner_enrollment/down.sql
	down004OwnerEnrollment string
)

func init() {
	registerSQLMigration(4, up004OwnerEnrollment, down004OwnerEnrollment)
}
