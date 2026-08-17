package management

import (
	"github.com/EonsofStupid/tessera/internal/domain"
	mgmt_pb "github.com/EonsofStupid/tessera/pkg/grpc/management"
)

func AddPasswordComplexityPolicyToDomain(req *mgmt_pb.AddCustomPasswordComplexityPolicyRequest) *domain.PasswordComplexityPolicy {
	return &domain.PasswordComplexityPolicy{
		MinLength:    req.MinLength,
		HasLowercase: req.HasLowercase,
		HasUppercase: req.HasUppercase,
		HasNumber:    req.HasNumber,
		HasSymbol:    req.HasSymbol,
	}
}

func UpdatePasswordComplexityPolicyToDomain(req *mgmt_pb.UpdateCustomPasswordComplexityPolicyRequest) *domain.PasswordComplexityPolicy {
	return &domain.PasswordComplexityPolicy{
		MinLength:    req.MinLength,
		HasLowercase: req.HasLowercase,
		HasUppercase: req.HasUppercase,
		HasNumber:    req.HasNumber,
		HasSymbol:    req.HasSymbol,
	}
}
