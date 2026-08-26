package domain_test

import (
	"os"
	"testing"

	"github.com/shippinAI/nomen/passwap"

	"github.com/shippinAI/nomen/backend/v3/domain"
	"github.com/shippinAI/nomen/internal/crypto"
)

func TestMain(m *testing.M) {
	os.Exit(func() int {
		domain.SetPasswordHasher(&crypto.Hasher{Swapper: &passwap.Swapper{}})
		return m.Run()
	}())
}
