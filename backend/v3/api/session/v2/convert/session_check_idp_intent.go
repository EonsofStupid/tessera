package convert

import (
	"github.com/shippinAI/nomen/backend/v3/domain"
	session_grpc "github.com/shippinAI/nomen/pkg/grpc/session/v2"
)

func CheckIDPIntentGRPCToDomain(checkIDPIntent *session_grpc.CheckIDPIntent) *domain.CheckIDPIntentType {
	if checkIDPIntent == nil {
		return nil
	}

	return &domain.CheckIDPIntentType{
		ID:    checkIDPIntent.GetIdpIntentId(),
		Token: checkIDPIntent.GetIdpIntentToken(),
	}
}
