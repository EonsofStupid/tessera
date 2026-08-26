package convert

import (
	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/query"
	"github.com/shippinAI/nomen/pkg/grpc/user/v2"
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
