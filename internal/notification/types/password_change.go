package types

import (
	"context"

	http_utils "github.com/shippinAI/nomen/internal/api/http"
	"github.com/shippinAI/nomen/internal/api/ui/console"
	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/query"
)

func (notify Notify) SendPasswordChange(ctx context.Context, user *query.NotifyUser) error {
	url := console.LoginHintLink(http_utils.DomainContext(ctx).Origin(), user.PreferredLoginName)
	args := make(map[string]interface{})
	return notify(url, args, domain.PasswordChangeMessageType, true)
}
