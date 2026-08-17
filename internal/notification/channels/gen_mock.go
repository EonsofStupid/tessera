package channels

//go:generate mockgen -package mock -destination ./mock/channel.mock.go github.com/EonsofStupid/tessera/internal/notification/channels NotificationChannel
//go:generate mockgen -package mock -destination ./mock/message.mock.go github.com/EonsofStupid/tessera/internal/notification/channels Message
