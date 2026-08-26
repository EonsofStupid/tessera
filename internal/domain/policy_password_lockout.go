package domain

import (
	"github.com/shippinAI/nomen/internal/eventstore/v1/models"
)

type LockoutPolicy struct {
	models.ObjectRoot

	Default             bool
	MaxPasswordAttempts uint64
	MaxOTPAttempts      uint64
	ShowLockOutFailures bool
}
