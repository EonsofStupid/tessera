package auth

import (
	"context"

	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/pkg/grpc/auth"
)

func UpdateMyEmailToDomain(ctx context.Context, email *auth.SetMyEmailRequest) *domain.Email {
	return &domain.Email{
		ObjectRoot:   ctxToObjectRoot(ctx),
		EmailAddress: domain.EmailAddress(email.Email),
	}
}
