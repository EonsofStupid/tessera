package middleware

import (
	"context"
	"slices"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/shippinAI/nomen/internal/activity"
	"github.com/shippinAI/nomen/internal/api/grpc/gerrors"
	ainfo "github.com/shippinAI/nomen/internal/api/info"
)

func ActivityInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = activityInfoFromGateway(ctx).SetMethod(info.FullMethod).IntoContext(ctx)
		resp, err := handler(ctx, req)
		if isResourceAPI(info.FullMethod) {
			code, _, _ := gerrors.ExtractNOMENError(err)
			ctx = ainfo.ActivityInfoFromContext(ctx).SetGRPCStatus(code).IntoContext(ctx)
			activity.TriggerGRPCWithContext(ctx, activity.ResourceAPI)
		}
		return resp, err
	}
}

var resourcePrefixes = []string{
	"/nomen.management.v1.ManagementService/",
	"/nomen.admin.v1.AdminService/",
	"/nomen.user.v2.UserService/",
	"/nomen.settings.v2.SettingsService/",
	"/nomen.user.v2beta.UserService/",
	"/nomen.settings.v2beta.SettingsService/",
	"/nomen.auth.v1.AuthService/",
}

func isResourceAPI(method string) bool {
	return slices.ContainsFunc(resourcePrefixes, func(prefix string) bool {
		return strings.HasPrefix(method, prefix)
	})
}

func activityInfoFromGateway(ctx context.Context) *ainfo.ActivityInfo {
	info := ainfo.ActivityInfoFromContext(ctx)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return info
	}
	path := md.Get(activity.PathKey)
	if len(path) != 1 {
		return info
	}
	requestMethod := md.Get(activity.RequestMethodKey)
	if len(requestMethod) != 1 {
		return info
	}
	return info.SetPath(path[0]).SetRequestMethod(requestMethod[0])
}
