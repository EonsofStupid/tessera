package overview

import (
	"context"
	"errors"
	"testing"

	"github.com/shippinAI/nomen/backend/v3/storage/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type queryerStub struct {
	row       database.Row
	statement string
	args      []any
}

func (s *queryerStub) QueryRow(_ context.Context, statement string, args ...any) database.Row {
	s.statement = statement
	s.args = args
	return s.row
}

type aggregateRow struct {
	humans, agents, attachments, flows int64
	policies                           []string
	err                                error
}

func (r aggregateRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(*int64) = r.humans
	*dest[1].(*int64) = r.agents
	*dest[2].(*int64) = r.attachments
	*dest[3].(*int64) = r.flows
	*dest[4].(*[]string) = r.policies
	return nil
}

func TestSnapshotReadsOnlyInstanceScopedNomenFacts(t *testing.T) {
	pool := &queryerStub{row: aggregateRow{
		humans: 2, agents: 3, attachments: 4, flows: 1,
		policies: []string{"policy-a"},
	}}
	facts, err := NewRepository(pool).Snapshot(context.Background(), "instance-1")

	require.NoError(t, err)
	assert.Equal(t, uint64(2), facts.HumanSeats)
	assert.Equal(t, uint64(3), facts.AgentSeats)
	assert.Equal(t, uint64(4), facts.WorkspaceAttachments)
	assert.Equal(t, uint64(1), facts.Flows)
	assert.Equal(t, []any{"instance-1"}, pool.args)
	assert.Contains(t, pool.statement, "FROM nomen_product.seats")
	assert.Contains(t, pool.statement, "FROM nomen_product.seat_workspaces")
	assert.Contains(t, pool.statement, "FROM nomen_product.flows")
	assert.NotContains(t, pool.statement, "billing")
}

func TestSnapshotDoesNotPromoteReadFailuresToEmptyFacts(t *testing.T) {
	pool := &queryerStub{row: aggregateRow{err: errors.New("database offline")}}
	_, err := NewRepository(pool).Snapshot(context.Background(), "instance-1")

	require.ErrorContains(t, err, "read overview facts")
}

func TestSnapshotRequiresInstance(t *testing.T) {
	_, err := NewRepository(&queryerStub{}).Snapshot(context.Background(), "")
	require.ErrorContains(t, err, "requires an instance")
}
