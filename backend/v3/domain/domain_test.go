package domain_test

import (
	"os"
	"testing"

	"github.com/zitadel/passwap"

	"github.com/EonsofStupid/tessera/backend/v3/domain"
	"github.com/EonsofStupid/tessera/internal/crypto"
)

func TestMain(m *testing.M) {
	os.Exit(func() int {
		domain.SetPasswordHasher(&crypto.Hasher{Swapper: &passwap.Swapper{}})
		return m.Run()
	}())
}
