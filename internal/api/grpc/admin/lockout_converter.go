package admin

import (
	"github.com/EonsofStupid/tessera/internal/domain"
	"github.com/EonsofStupid/tessera/pkg/grpc/admin"
)

func UpdateLockoutPolicyToDomain(p *admin.UpdateLockoutPolicyRequest) *domain.LockoutPolicy {
	return &domain.LockoutPolicy{
		MaxPasswordAttempts: uint64(p.MaxPasswordAttempts),
		MaxOTPAttempts:      uint64(p.MaxOtpAttempts),
	}
}
