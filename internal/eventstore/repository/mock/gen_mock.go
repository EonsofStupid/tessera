package mock

//go:generate mockgen -package mock -destination ./repository.mock.go github.com/shippinAI/nomen/internal/eventstore Querier,Pusher
