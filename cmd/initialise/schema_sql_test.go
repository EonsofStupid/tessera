package initialise

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitSQLCreatesEveryNomenSchema(t *testing.T) {
	t.Parallel()
	require.NoError(t, ReadStmts())
	require.Contains(t, createEventstoreStmt, "CREATE SCHEMA IF NOT EXISTS eventstore")
	require.Contains(t, createProjectionsStmt, "CREATE SCHEMA IF NOT EXISTS projections")
	require.Contains(t, createSystemStmt, "CREATE SCHEMA IF NOT EXISTS system")
	require.Contains(t, createNomenSchemaStmt, "CREATE SCHEMA IF NOT EXISTS nomen")
	require.Contains(t, createNomenProductSchemaStmt, "CREATE SCHEMA IF NOT EXISTS nomen_product")
}
