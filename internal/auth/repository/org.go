package repository

import (
	"context"

	"github.com/EonsofStupid/tessera/internal/domain"
	iam_model "github.com/EonsofStupid/tessera/internal/iam/model"
)

type OrgRepository interface {
	GetMyPasswordComplexityPolicy(ctx context.Context) (*iam_model.PasswordComplexityPolicyView, error)
	GetLoginText(ctx context.Context, orgID string) ([]*domain.CustomText, error)
}
