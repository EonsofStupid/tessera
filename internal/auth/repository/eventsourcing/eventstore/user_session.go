package eventstore

import (
	"context"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/auth/repository/eventsourcing/view"
	usr_model "github.com/shippinAI/nomen/internal/user/model"
	"github.com/shippinAI/nomen/internal/user/repository/view/model"
)

type UserSessionRepo struct {
	View *view.View
}

func (repo *UserSessionRepo) GetMyUserSessions(ctx context.Context) ([]*usr_model.UserSessionView, error) {
	userSessions, err := repo.View.UserSessionsByAgentID(ctx, authz.GetCtxData(ctx).AgentID, authz.GetInstance(ctx).InstanceID())
	if err != nil {
		return nil, err
	}
	return model.UserSessionsToModel(userSessions), nil
}
