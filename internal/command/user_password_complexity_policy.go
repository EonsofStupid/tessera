package command

import (
	"context"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/command/preparation"
	"github.com/shippinAI/nomen/internal/zerrors"
)

func passwordComplexityPolicyWriteModel(ctx context.Context, filter preparation.FilterToQueryReducer) (*PasswordComplexityPolicyWriteModel, error) {
	wm, err := customPasswordComplexityPolicy(ctx, filter)
	if err != nil || wm != nil && wm.State.Exists() {
		return wm, err
	}
	wm, err = defaultPasswordComplexityPolicy(ctx, filter)
	if err != nil || wm != nil {
		return wm, err
	}
	return nil, zerrors.ThrowInternal(nil, "USER-uQ96e", "Errors.Internal")
}

func customPasswordComplexityPolicy(ctx context.Context, filter preparation.FilterToQueryReducer) (*PasswordComplexityPolicyWriteModel, error) {
	policy := NewOrgPasswordComplexityPolicyWriteModel(authz.GetCtxData(ctx).OrgID)
	events, err := filter(ctx, policy.Query())
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}
	policy.AppendEvents(events...)
	err = policy.Reduce()
	return &policy.PasswordComplexityPolicyWriteModel, err
}

func defaultPasswordComplexityPolicy(ctx context.Context, filter preparation.FilterToQueryReducer) (*PasswordComplexityPolicyWriteModel, error) {
	policy := NewInstancePasswordComplexityPolicyWriteModel(ctx)
	events, err := filter(ctx, policy.Query())
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}
	policy.AppendEvents(events...)
	err = policy.Reduce()
	return &policy.PasswordComplexityPolicyWriteModel, err
}
