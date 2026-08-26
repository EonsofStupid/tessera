package user

import (
	"github.com/shippinAI/nomen/internal/api/grpc/object"
	"github.com/shippinAI/nomen/internal/domain"
	user_model "github.com/shippinAI/nomen/internal/user/model"
	"github.com/shippinAI/nomen/pkg/grpc/user"
)

func UserSessionsToPb(sessions []*user_model.UserSessionView, avatarPrefix string) []*user.Session {
	s := make([]*user.Session, len(sessions))
	for i, session := range sessions {
		s[i] = UserSessionToPb(session, avatarPrefix)
	}
	return s
}

func UserSessionToPb(session *user_model.UserSessionView, avatarPrefix string) *user.Session {
	return &user.Session{
		// SessionId: session.,//TOOD: not return from be
		AgentId:     session.UserAgentID,
		UserId:      session.UserID,
		UserName:    session.UserName,
		LoginName:   session.LoginName,
		DisplayName: session.DisplayName,
		AuthState:   SessionStateToPb(session.State),
		AvatarUrl:   domain.AvatarURL(avatarPrefix, session.ResourceOwner, session.AvatarKey),
		Details: object.ToViewDetailsPb(
			session.Sequence,
			session.CreationDate,
			session.ChangeDate,
			session.ResourceOwner,
		),
	}
}

func SessionStateToPb(state domain.UserSessionState) user.SessionState {
	switch state {
	case domain.UserSessionStateActive:
		return user.SessionState_SESSION_STATE_ACTIVE
	case domain.UserSessionStateTerminated:
		return user.SessionState_SESSION_STATE_TERMINATED
	default:
		return user.SessionState_SESSION_STATE_UNSPECIFIED
	}
}
