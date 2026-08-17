package convert

import (
	"github.com/EonsofStupid/tessera/backend/v3/domain"
	session_grpc "github.com/EonsofStupid/tessera/pkg/grpc/session/v2"
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
