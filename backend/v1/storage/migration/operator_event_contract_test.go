package migration

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOperatorEventMigrationUsesPrefixedForcedRLSTables(t *testing.T) {
	t.Parallel()
	for _, table := range []string{"nomen_operator_events", "nomen_outbox"} {
		require.Contains(t, up003OperatorEvents, "CREATE TABLE nomen_product."+table)
		require.Contains(t, up003OperatorEvents, "ALTER TABLE nomen_product."+table+" ENABLE ROW LEVEL SECURITY")
		require.Contains(t, up003OperatorEvents, "ALTER TABLE nomen_product."+table+" FORCE ROW LEVEL SECURITY")
	}
	for _, match := range regexp.MustCompile(`CREATE TABLE nomen\.([a-z0-9_]+)`).FindAllStringSubmatch(up003OperatorEvents, -1) {
		require.True(t, strings.HasPrefix(match[1], "nomen_"), "new table %s must use the nomen_ prefix", match[1])
	}
	require.Contains(t, up003OperatorEvents, "current_setting('nomen_product.tenant_id', true)")
	require.Contains(t, up003OperatorEvents, "current_setting('nomen_product.instance_id', true)")
}
