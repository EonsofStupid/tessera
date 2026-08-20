package integration

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegrationPrivateKeyIsGeneratedWithoutARepositoryFixture(t *testing.T) {
	t.Setenv("TESSERA_TEST_GENERATED_KEY_FILE", "")
	key, err := integrationPrivateKey("TESSERA_TEST_GENERATED_KEY_FILE")
	require.NoError(t, err)
	block, _ := pem.Decode(key)
	require.NotNil(t, block)
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	require.NoError(t, err)
	_, ok := parsed.(*rsa.PrivateKey)
	require.True(t, ok)
}

func TestIntegrationPrivateKeyReadsOnlyAnExplicitProtectedPath(t *testing.T) {
	t.Setenv("TESSERA_TEST_MISSING_KEY_FILE", "/path/that/does/not/exist")
	_, err := integrationPrivateKey("TESSERA_TEST_MISSING_KEY_FILE")
	require.ErrorContains(t, err, "TESSERA_TEST_MISSING_KEY_FILE")
	require.True(t, errors.Is(err, os.ErrNotExist))
}
