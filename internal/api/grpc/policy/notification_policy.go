package policy

import (
	"github.com/EonsofStupid/tessera/internal/api/grpc/object"
	"github.com/EonsofStupid/tessera/internal/query"
	policy_pb "github.com/EonsofStupid/tessera/pkg/grpc/policy"
)

func ModelNotificationPolicyToPb(policy *query.NotificationPolicy) *policy_pb.NotificationPolicy {
	return &policy_pb.NotificationPolicy{
		IsDefault:      policy.IsDefault,
		PasswordChange: policy.PasswordChange,
		Details: object.ToViewDetailsPb(
			policy.Sequence,
			policy.CreationDate,
			policy.ChangeDate,
			policy.ResourceOwner,
		),
	}
}
