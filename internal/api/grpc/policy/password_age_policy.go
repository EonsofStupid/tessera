package policy

import (
	"github.com/EonsofStupid/tessera/internal/api/grpc/object"
	"github.com/EonsofStupid/tessera/internal/query"
	policy_pb "github.com/EonsofStupid/tessera/pkg/grpc/policy"
)

func ModelPasswordAgePolicyToPb(policy *query.PasswordAgePolicy) *policy_pb.PasswordAgePolicy {
	return &policy_pb.PasswordAgePolicy{
		IsDefault:      policy.IsDefault,
		MaxAgeDays:     policy.MaxAgeDays,
		ExpireWarnDays: policy.ExpireWarnDays,
		Details: object.ToViewDetailsPb(
			policy.Sequence,
			policy.CreationDate,
			policy.ChangeDate,
			policy.ResourceOwner,
		),
	}
}
