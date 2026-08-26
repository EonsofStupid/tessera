package login

import (
	"context"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/domain"
)

func (l *Login) bulkSetUserMetadata(ctx context.Context, userID, orgID string, metadata []*domain.Metadata) error {
	// user context necessary due to permission check in command
	userCtx := authz.SetCtxData(ctx, authz.CtxData{UserID: userID, OrgID: orgID})
	_, err := l.command.BulkSetUserMetadata(userCtx, userID, orgID, nil, metadata...)
	return err
}
