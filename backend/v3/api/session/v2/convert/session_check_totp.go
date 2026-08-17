package convert

import (
	"github.com/EonsofStupid/tessera/backend/v3/domain"
	session_grpc "github.com/EonsofStupid/tessera/pkg/grpc/session/v2"
)

func CheckTOTPGRPCToDomain(checkTOTP *session_grpc.CheckTOTP) *domain.CheckTOTPType {
	if checkTOTP == nil {
		return nil
	}

	return &domain.CheckTOTPType{
		Code: checkTOTP.GetCode(),
	}
}
