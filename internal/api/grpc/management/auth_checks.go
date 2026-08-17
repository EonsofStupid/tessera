package management

import (
	"context"

	"github.com/EonsofStupid/tessera/internal/api/authz"
	"github.com/EonsofStupid/tessera/internal/command"
	"github.com/EonsofStupid/tessera/internal/zerrors"
)

func checkExplicitProjectPermission(ctx context.Context) command.UserGrantPermissionCheck {
	permissions := authz.GetRequestPermissionsFromCtx(ctx)
	if authz.HasGlobalPermission(permissions) {
		return nil
	}
	ids := authz.GetAllPermissionCtxIDs(permissions)
	return func(projectID, grantID string) command.PermissionCheck {
		return func(resourceOwner, aggregateID string) error {
			if grantID != "" && listContainsID(ids, grantID) {
				return nil
			}
			if listContainsID(ids, projectID) {
				return nil
			}
			return zerrors.ThrowPermissionDenied(nil, "EVENT-Shu7e", "Errors.UserGrant.NoPermissionForProject")
		}
	}
}

func listContainsID(ids []string, id string) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}
