package setup

import (
	"context"
	_ "embed"

	"github.com/shippinAI/nomen/backend/v3/instrumentation/logging"
	"github.com/shippinAI/nomen/internal/database"
	"github.com/shippinAI/nomen/internal/eventstore"
)

var (
	//go:embed 16.sql
	uniqueConstraintLower string
)

type UniqueConstraintToLower struct {
	dbClient *database.DB
}

func (mig *UniqueConstraintToLower) Execute(ctx context.Context, _ eventstore.Event) error {
	res, err := mig.dbClient.ExecContext(ctx, uniqueConstraintLower)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	logging.Info(ctx, "unique constraints updated", "count", count)
	return err
}

func (mig *UniqueConstraintToLower) String() string {
	return "16_unique_constraint_lower"
}
