package admin

import (
	"github.com/EonsofStupid/tessera/internal/domain"
	admin_pb "github.com/EonsofStupid/tessera/pkg/grpc/admin"
)

func UpdatePasswordAgePolicyToDomain(policy *admin_pb.UpdatePasswordAgePolicyRequest) *domain.PasswordAgePolicy {
	return &domain.PasswordAgePolicy{
		MaxAgeDays:     uint64(policy.MaxAgeDays),
		ExpireWarnDays: uint64(policy.ExpireWarnDays),
	}
}
