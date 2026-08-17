package auth

import (
	"context"

	"github.com/EonsofStupid/tessera/internal/domain"
	"github.com/EonsofStupid/tessera/pkg/grpc/auth"
)

func UpdateMyEmailToDomain(ctx context.Context, email *auth.SetMyEmailRequest) *domain.Email {
	return &domain.Email{
		ObjectRoot:   ctxToObjectRoot(ctx),
		EmailAddress: domain.EmailAddress(email.Email),
	}
}
