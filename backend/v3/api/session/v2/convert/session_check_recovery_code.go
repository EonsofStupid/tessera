package convert

import (
	"github.com/shippinAI/nomen/backend/v3/domain"
	session_grpc "github.com/shippinAI/nomen/pkg/grpc/session/v2"
)

func CheckRecoveryCodeGRPCToDomain(checkRecoveryCode *session_grpc.CheckRecoveryCode) *domain.CheckTypeRecoveryCode {
	if checkRecoveryCode == nil {
		return nil
	}
	return &domain.CheckTypeRecoveryCode{
		RecoveryCode: checkRecoveryCode.GetCode(),
	}
}
