package migration

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOperatorEventMigrationUsesPrefixedForcedRLSTables(t *testing.T) {
	t.Parallel()
	for _, table := range []string{"tessera_operator_events", "tessera_outbox"} {
		require.Contains(t, up003OperatorEvents, "CREATE TABLE tessera."+table)
		require.Contains(t, up003OperatorEvents, "ALTER TABLE tessera."+table+" ENABLE ROW LEVEL SECURITY")
		require.Contains(t, up003OperatorEvents, "ALTER TABLE tessera."+table+" FORCE ROW LEVEL SECURITY")
	}
	for _, match := range regexp.MustCompile(`CREATE TABLE tessera\.([a-z0-9_]+)`).FindAllStringSubmatch(up003OperatorEvents, -1) {
		require.True(t, strings.HasPrefix(match[1], "tessera_"), "new table %s must use the tessera_ prefix", match[1])
	}
	require.Contains(t, up003OperatorEvents, "current_setting('tessera.tenant_id', true)")
	require.Contains(t, up003OperatorEvents, "current_setting('tessera.instance_id', true)")
}
