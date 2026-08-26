package oidc

import (
	"context"

	"github.com/shippinAI/nomen/oidc/v3/pkg/op"
)

func (o *OPStorage) JWTProfileTokenType(context.Context, op.TokenRequest) (op.AccessTokenType, error) {
	panic(o.panicErr("JWTProfileTokenType"))
}
