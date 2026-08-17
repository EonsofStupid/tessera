package setup

import (
	"context"
	_ "embed"

	"github.com/EonsofStupid/tessera/internal/database"
	"github.com/EonsofStupid/tessera/internal/eventstore"
)

var (
	//go:embed 75.sql
	addOIDCAppLinkConfig string
)

type Apps7OIDCConfigsAddAppLinkConfig struct {
	dbClient *database.DB
}

func (mig *Apps7OIDCConfigsAddAppLinkConfig) Execute(ctx context.Context, _ eventstore.Event) error {
	_, err := mig.dbClient.ExecContext(ctx, addOIDCAppLinkConfig)
	return err
}

func (mig *Apps7OIDCConfigsAddAppLinkConfig) String() string {
	return "75_apps7_oidc_configs_add_app_link_config"
}
