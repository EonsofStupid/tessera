package migration

import _ "embed"

var (
	//go:embed 005_owner_passkey/up.sql
	up005OwnerPasskey string
	//go:embed 005_owner_passkey/down.sql
	down005OwnerPasskey string
)

func init() {
	registerSQLMigration(5, up005OwnerPasskey, down005OwnerPasskey)
}
