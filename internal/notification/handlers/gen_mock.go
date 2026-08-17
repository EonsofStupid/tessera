package handlers

//go:generate mockgen -package mock -destination ./mock/queries.mock.go github.com/EonsofStupid/tessera/internal/notification/handlers Queries
//go:generate mockgen -package mock -destination ./mock/commands.mock.go github.com/EonsofStupid/tessera/internal/notification/handlers Commands
//go:generate mockgen -package mock -destination ./mock/queue.mock.go github.com/EonsofStupid/tessera/internal/notification/handlers Queue
