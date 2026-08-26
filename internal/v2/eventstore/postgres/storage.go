package postgres

import (
	"context"

	"github.com/shippinAI/nomen/internal/database"
	"github.com/shippinAI/nomen/internal/v2/eventstore"
)

var (
	_ eventstore.Pusher  = (*Storage)(nil)
	_ eventstore.Querier = (*Storage)(nil)
)

type Storage struct {
	client *database.DB
	config *Config
}

type Config struct {
	MaxRetries uint32
}

func New(client *database.DB, config *Config) *Storage {
	return &Storage{
		client: client,
		config: config,
	}
}

// Health implements eventstore.Pusher.
func (s *Storage) Health(ctx context.Context) error {
	return s.client.PingContext(ctx)
}
