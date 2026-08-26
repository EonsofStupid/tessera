package config

import (
	"database/sql"

	"github.com/shippinAI/nomen/internal/api/http/middleware"
	"github.com/shippinAI/nomen/internal/static"
	"github.com/shippinAI/nomen/internal/static/database"
	"github.com/shippinAI/nomen/internal/static/s3"
	"github.com/shippinAI/nomen/internal/zerrors"
)

type AssetStorageConfig struct {
	Type   string
	Cache  middleware.CacheConfig
	Config map[string]interface{} `mapstructure:",remain"`
}

func (a *AssetStorageConfig) NewStorage(client *sql.DB) (static.Storage, error) {
	t, ok := storage[a.Type]
	if !ok {
		return nil, zerrors.ThrowInternalf(nil, "STATIC-dsbjh", "config type %s not supported", a.Type)
	}

	return t(client, a.Config)
}

var storage = map[string]static.CreateStorage{
	"db": database.NewStorage,
	"":   database.NewStorage,
	"s3": s3.NewStorage,
}
