package setup

import (
	"context"
	_ "embed"

	"github.com/shippinAI/nomen/internal/database"
	"github.com/shippinAI/nomen/internal/eventstore"
)

var (
	//go:embed 13/13_fix_quota_constraints.sql
	fixQuotaConstraints string
)

type FixQuotaConstraints struct {
	dbClient *database.DB
}

func (mig *FixQuotaConstraints) Execute(ctx context.Context, _ eventstore.Event) error {
	_, err := mig.dbClient.ExecContext(ctx, fixQuotaConstraints)
	return err
}

func (mig *FixQuotaConstraints) String() string {
	return "13_fix_quota_constraints"
}
