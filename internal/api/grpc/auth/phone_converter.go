package auth

import (
	"context"

	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/pkg/grpc/auth"
)

func UpdateMyPhoneToDomain(ctx context.Context, phone *auth.SetMyPhoneRequest) *domain.Phone {
	return &domain.Phone{
		ObjectRoot:  ctxToObjectRoot(ctx),
		PhoneNumber: domain.PhoneNumber(phone.Phone),
	}
}
