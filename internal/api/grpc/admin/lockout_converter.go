package admin

import (
	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/pkg/grpc/admin"
)

func UpdateLockoutPolicyToDomain(p *admin.UpdateLockoutPolicyRequest) *domain.LockoutPolicy {
	return &domain.LockoutPolicy{
		MaxPasswordAttempts: uint64(p.MaxPasswordAttempts),
		MaxOTPAttempts:      uint64(p.MaxOtpAttempts),
	}
}
