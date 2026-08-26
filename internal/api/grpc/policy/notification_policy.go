package policy

import (
	"github.com/shippinAI/nomen/internal/api/grpc/object"
	"github.com/shippinAI/nomen/internal/query"
	policy_pb "github.com/shippinAI/nomen/pkg/grpc/policy"
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
