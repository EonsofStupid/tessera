package mock

//go:generate mockgen -package mock -destination ./repository.mock.go github.com/EonsofStupid/tessera/internal/eventstore Querier,Pusher
