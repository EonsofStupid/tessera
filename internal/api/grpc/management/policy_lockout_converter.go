package management

import (
	"github.com/EonsofStupid/tessera/internal/domain"
	mgmt "github.com/EonsofStupid/tessera/pkg/grpc/management"
)

func AddLockoutPolicyToDomain(p *mgmt.AddCustomLockoutPolicyRequest) *domain.LockoutPolicy {
	return &domain.LockoutPolicy{
		MaxPasswordAttempts: uint64(p.MaxPasswordAttempts),
		MaxOTPAttempts:      uint64(p.MaxOtpAttempts),
	}
}

func UpdateLockoutPolicyToDomain(p *mgmt.UpdateCustomLockoutPolicyRequest) *domain.LockoutPolicy {
	return &domain.LockoutPolicy{
		MaxPasswordAttempts: uint64(p.MaxPasswordAttempts),
		MaxOTPAttempts:      uint64(p.MaxOtpAttempts),
	}
}
