package convert

import (
	"github.com/EonsofStupid/tessera/backend/v3/domain"
	session_grpc "github.com/EonsofStupid/tessera/pkg/grpc/session/v2"
)

func CheckPasswordGRPCToDomain(checkPsw *session_grpc.CheckPassword) *domain.CheckPasswordType {
	if checkPsw == nil {
		return nil
	}

	return &domain.CheckPasswordType{
		Password: checkPsw.GetPassword(),
	}
}
