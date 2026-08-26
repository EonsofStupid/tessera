//go:build integration

package management_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/shippinAI/nomen/internal/integration"
	mgmt_pb "github.com/shippinAI/nomen/pkg/grpc/management"
)

var (
	CTX, IAMOwnerCTX, OrgCTX context.Context
	Instance                 *integration.Instance
	Client                   mgmt_pb.ManagementServiceClient
)

func TestMain(m *testing.M) {
	os.Exit(func() int {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		Instance = integration.NewInstance(ctx)
		CTX = ctx
		IAMOwnerCTX = Instance.WithAuthorization(ctx, integration.UserTypeIAMOwner)
		OrgCTX = Instance.WithAuthorization(ctx, integration.UserTypeOrgOwner)
		Client = Instance.Client.Mgmt
		return m.Run()
	}())
}
