package admin

import (
	"github.com/shippinAI/nomen/internal/domain"
	admin_pb "github.com/shippinAI/nomen/pkg/grpc/admin"
)

func UpdatePasswordAgePolicyToDomain(policy *admin_pb.UpdatePasswordAgePolicyRequest) *domain.PasswordAgePolicy {
	return &domain.PasswordAgePolicy{
		MaxAgeDays:     uint64(policy.MaxAgeDays),
		ExpireWarnDays: uint64(policy.ExpireWarnDays),
	}
}
