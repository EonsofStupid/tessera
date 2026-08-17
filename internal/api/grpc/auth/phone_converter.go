package auth

import (
	"context"

	"github.com/EonsofStupid/tessera/internal/domain"
	"github.com/EonsofStupid/tessera/pkg/grpc/auth"
)

func UpdateMyPhoneToDomain(ctx context.Context, phone *auth.SetMyPhoneRequest) *domain.Phone {
	return &domain.Phone{
		ObjectRoot:  ctxToObjectRoot(ctx),
		PhoneNumber: domain.PhoneNumber(phone.Phone),
	}
}
