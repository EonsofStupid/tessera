package gomap

import (
	"github.com/EonsofStupid/tessera/backend/v3/storage/cache"
)

type Config struct {
	Enabled   bool
	AutoPrune cache.AutoPruneConfig
}

type Connector struct {
	Config cache.AutoPruneConfig
}

func NewConnector(config Config) *Connector {
	if !config.Enabled {
		return nil
	}
	return &Connector{
		Config: config.AutoPrune,
	}
}
