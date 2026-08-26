package migration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOwnerEnrollmentMigrationIsInstanceScopedAndStoresDigestsOnly(t *testing.T) {
	t.Parallel()
	require.Contains(t, up004OwnerEnrollment, "CREATE TABLE nomen_product.nomen_owner_enrollments")
	require.Contains(t, up004OwnerEnrollment, "ENABLE ROW LEVEL SECURITY")
	require.Contains(t, up004OwnerEnrollment, "FORCE ROW LEVEL SECURITY")
	require.Contains(t, up004OwnerEnrollment, "current_setting('nomen_product.instance_id', true)")
	require.Contains(t, up004OwnerEnrollment, "challenge_digest")
	require.Contains(t, up004OwnerEnrollment, "idempotency_key_digest")
	require.NotContains(t, up004OwnerEnrollment, "bootstrap_authority")
	require.NotContains(t, up004OwnerEnrollment, "private_key")
}
