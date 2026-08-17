package repository

import (
	"context"

	"github.com/EonsofStupid/tessera/internal/user/model"
)

type UserSessionRepository interface {
	GetMyUserSessions(ctx context.Context) ([]*model.UserSessionView, error)
}
