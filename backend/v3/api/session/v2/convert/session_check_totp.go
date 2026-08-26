package convert

import (
	"github.com/shippinAI/nomen/backend/v3/domain"
	session_grpc "github.com/shippinAI/nomen/pkg/grpc/session/v2"
)

func CheckTOTPGRPCToDomain(checkTOTP *session_grpc.CheckTOTP) *domain.CheckTOTPType {
	if checkTOTP == nil {
		return nil
	}

	return &domain.CheckTOTPType{
		Code: checkTOTP.GetCode(),
	}
}
