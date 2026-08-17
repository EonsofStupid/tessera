package management

import (
	"github.com/EonsofStupid/tessera/internal/domain"
	mgmt_pb "github.com/EonsofStupid/tessera/pkg/grpc/management"
)

func AddPasswordAgePolicyToDomain(policy *mgmt_pb.AddCustomPasswordAgePolicyRequest) *domain.PasswordAgePolicy {
	return &domain.PasswordAgePolicy{
		MaxAgeDays:     uint64(policy.MaxAgeDays),
		ExpireWarnDays: uint64(policy.ExpireWarnDays),
	}
}

func UpdatePasswordAgePolicyToDomain(policy *mgmt_pb.UpdateCustomPasswordAgePolicyRequest) *domain.PasswordAgePolicy {
	return &domain.PasswordAgePolicy{
		MaxAgeDays:     uint64(policy.MaxAgeDays),
		ExpireWarnDays: uint64(policy.ExpireWarnDays),
	}
}
