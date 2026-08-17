package policy

import (
	"github.com/EonsofStupid/tessera/internal/api/grpc/object"
	"github.com/EonsofStupid/tessera/internal/query"
	policy_pb "github.com/EonsofStupid/tessera/pkg/grpc/policy"
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
