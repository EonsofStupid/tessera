package convert

import (
	"github.com/EonsofStupid/tessera/internal/domain"
	"github.com/EonsofStupid/tessera/internal/query"
	"github.com/EonsofStupid/tessera/pkg/grpc/user/v2"
)

func machineToPb(userQ *query.Machine) *user.MachineUser {
	return &user.MachineUser{
		Name:            userQ.Name,
		Description:     userQ.Description,
		HasSecret:       userQ.EncodedSecret != "",
		AccessTokenType: accessTokenTypeToPb(userQ.AccessTokenType),
	}
}

func accessTokenTypeToPb(accessTokenType domain.OIDCTokenType) user.AccessTokenType {
	switch accessTokenType {
	case domain.OIDCTokenTypeBearer:
		return user.AccessTokenType_ACCESS_TOKEN_TYPE_BEARER
	case domain.OIDCTokenTypeJWT:
		return user.AccessTokenType_ACCESS_TOKEN_TYPE_JWT
	default:
		return user.AccessTokenType_ACCESS_TOKEN_TYPE_BEARER
	}
}
