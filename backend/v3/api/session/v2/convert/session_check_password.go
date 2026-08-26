package convert

import (
	"github.com/shippinAI/nomen/backend/v3/domain"
	session_grpc "github.com/shippinAI/nomen/pkg/grpc/session/v2"
)

func CheckPasswordGRPCToDomain(checkPsw *session_grpc.CheckPassword) *domain.CheckPasswordType {
	if checkPsw == nil {
		return nil
	}

	return &domain.CheckPasswordType{
		Password: checkPsw.GetPassword(),
	}
}
