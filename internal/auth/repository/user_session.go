package repository

import (
	"context"

	"github.com/shippinAI/nomen/internal/user/model"
)

type UserSessionRepository interface {
	GetMyUserSessions(ctx context.Context) ([]*model.UserSessionView, error)
}
