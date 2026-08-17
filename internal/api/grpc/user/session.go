package user

import (
	"github.com/EonsofStupid/tessera/internal/api/grpc/object"
	"github.com/EonsofStupid/tessera/internal/domain"
	user_model "github.com/EonsofStupid/tessera/internal/user/model"
	"github.com/EonsofStupid/tessera/pkg/grpc/user"
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
