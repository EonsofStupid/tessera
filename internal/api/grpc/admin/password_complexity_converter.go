package admin

import (
	"github.com/EonsofStupid/tessera/internal/domain"
	admin_pb "github.com/EonsofStupid/tessera/pkg/grpc/admin"
)

func UpdatePasswordComplexityPolicyToDomain(req *admin_pb.UpdatePasswordComplexityPolicyRequest) *domain.PasswordComplexityPolicy {
	return &domain.PasswordComplexityPolicy{
		MinLength:    uint64(req.MinLength),
		HasLowercase: req.HasLowercase,
		HasUppercase: req.HasUppercase,
		HasNumber:    req.HasNumber,
		HasSymbol:    req.HasSymbol,
	}
}
