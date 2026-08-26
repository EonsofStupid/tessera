package policy

import (
	"github.com/shippinAI/nomen/internal/api/grpc/object"
	"github.com/shippinAI/nomen/internal/query"
	policy_pb "github.com/shippinAI/nomen/pkg/grpc/policy"
)

func ModelPasswordComplexityPolicyToPb(policy *query.PasswordComplexityPolicy) *policy_pb.PasswordComplexityPolicy {
	return &policy_pb.PasswordComplexityPolicy{
		IsDefault:    policy.IsDefault,
		MinLength:    policy.MinLength,
		HasUppercase: policy.HasUppercase,
		HasLowercase: policy.HasLowercase,
		HasNumber:    policy.HasNumber,
		HasSymbol:    policy.HasSymbol,
		Details: object.ToViewDetailsPb(
			policy.Sequence,
			policy.CreationDate,
			policy.ChangeDate,
			policy.ResourceOwner,
		),
	}
}
