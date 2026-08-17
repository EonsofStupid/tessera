package types

import (
	"context"

	http_utils "github.com/EonsofStupid/tessera/internal/api/http"
	"github.com/EonsofStupid/tessera/internal/api/ui/console"
	"github.com/EonsofStupid/tessera/internal/domain"
	"github.com/EonsofStupid/tessera/internal/query"
)

func (notify Notify) SendPasswordChange(ctx context.Context, user *query.NotifyUser) error {
	url := console.LoginHintLink(http_utils.DomainContext(ctx).Origin(), user.PreferredLoginName)
	args := make(map[string]interface{})
	return notify(url, args, domain.PasswordChangeMessageType, true)
}
