package types

import (
	"context"
	"strings"

	http_utils "github.com/shippinAI/nomen/internal/api/http"
	"github.com/shippinAI/nomen/internal/api/ui/login"
	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/query"
)

func (notify Notify) SendInviteCode(ctx context.Context, user *query.NotifyUser, code, applicationName, urlTmpl, authRequestID string) error {
	var url string
	if applicationName == "" {
		applicationName = "NOMEN"
	}
	if urlTmpl == "" {
		url = login.InviteUserLink(http_utils.DomainContext(ctx).Origin(), user.ID, user.PreferredLoginName, code, user.ResourceOwner, authRequestID)
	} else {
		var buf strings.Builder
		if err := domain.RenderConfirmURLTemplate(&buf, urlTmpl, user.ID, code, user.ResourceOwner); err != nil {
			return err
		}
		url = buf.String()
	}
	args := make(map[string]interface{})
	args["Code"] = code
	args["ApplicationName"] = applicationName
	return notify(url, args, domain.InviteUserMessageType, true)
}
