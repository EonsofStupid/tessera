package eventsourcing

import (
	"context"

	admin_handler "github.com/shippinAI/nomen/internal/admin/repository/eventsourcing/handler"
	admin_view "github.com/shippinAI/nomen/internal/admin/repository/eventsourcing/view"
	"github.com/shippinAI/nomen/internal/database"
	"github.com/shippinAI/nomen/internal/query"
	"github.com/shippinAI/nomen/internal/static"
)

type Config struct {
	Spooler admin_handler.Config
}

func Start(ctx context.Context, conf Config, static static.Storage, dbClient *database.DB, queries *query.Queries) error {
	view, err := admin_view.StartView(dbClient)
	if err != nil {
		return err
	}

	admin_handler.Register(ctx, conf.Spooler, view, static)
	admin_handler.Start(ctx)

	return nil
}
