package policy

import (
	"github.com/shippinAI/nomen/internal/api/grpc/object"
	"github.com/shippinAI/nomen/internal/query"
	policy_pb "github.com/shippinAI/nomen/pkg/grpc/policy"
)

func ModelLockoutPolicyToPb(policy *query.LockoutPolicy) *policy_pb.LockoutPolicy {
	return &policy_pb.LockoutPolicy{
		IsDefault:           policy.IsDefault,
		MaxPasswordAttempts: policy.MaxPasswordAttempts,
		MaxOtpAttempts:      policy.MaxOTPAttempts,
		Details: object.ToViewDetailsPb(
			policy.Sequence,
			policy.CreationDate,
			policy.ChangeDate,
			policy.ResourceOwner,
		),
	}
}
