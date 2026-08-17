package types

import (
	"context"

	http_utils "github.com/EonsofStupid/tessera/internal/api/http"
	"github.com/EonsofStupid/tessera/internal/api/ui/login"
	"github.com/EonsofStupid/tessera/internal/domain"
	"github.com/EonsofStupid/tessera/internal/query"
)

func (notify Notify) SendUserInitCode(ctx context.Context, user *query.NotifyUser, code, authRequestID string) error {
	url := login.InitUserLink(http_utils.DomainContext(ctx).Origin(), user.ID, user.PreferredLoginName, code, user.ResourceOwner, user.PasswordSet, authRequestID)
	args := make(map[string]interface{})
	args["Code"] = code
	return notify(url, args, domain.InitCodeMessageType, true)
}
