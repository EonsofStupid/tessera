package convert

import (
	"github.com/EonsofStupid/tessera/backend/v3/domain"
	session_grpc "github.com/EonsofStupid/tessera/pkg/grpc/session/v2"
)

func CheckRecoveryCodeGRPCToDomain(checkRecoveryCode *session_grpc.CheckRecoveryCode) *domain.CheckTypeRecoveryCode {
	if checkRecoveryCode == nil {
		return nil
	}
	return &domain.CheckTypeRecoveryCode{
		RecoveryCode: checkRecoveryCode.GetCode(),
	}
}
