package pg

import (
	"github.com/shippinAI/nomen/internal/cache"
	"github.com/shippinAI/nomen/internal/database"
)

type Config struct {
	Enabled   bool
	AutoPrune cache.AutoPruneConfig
}

type Connector struct {
	PGXPool
	Config Config
}

func NewConnector(config Config, client *database.DB) *Connector {
	if !config.Enabled {
		return nil
	}
	return &Connector{
		PGXPool: client.Pool,
		Config:  config,
	}
}
