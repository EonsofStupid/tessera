package integration

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/zitadel/oidc/v3/pkg/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	http_util "github.com/EonsofStupid/tessera/internal/api/http"
	"github.com/EonsofStupid/tessera/pkg/grpc/system"
)

var systemUserKey = sync.OnceValues(func() ([]byte, error) {
	return integrationPrivateKey("TESSERA_INTEGRATION_SYSTEM_USER_KEY_FILE")
})

var systemUserWithNoPermissions = sync.OnceValues(func() ([]byte, error) {
	return integrationPrivateKey("TESSERA_INTEGRATION_UNPRIVILEGED_KEY_FILE")
})

var (
	// SystemClient creates a system connection once and reuses it on every use.
	// Each client call automatically gets the authorization context for the system user.
	SystemClient                     = sync.OnceValue[system.SystemServiceClient](systemClient)
	SystemToken                      string
	SystemUserWithNoPermissionsToken string
)

func systemClient() system.SystemServiceClient {
	cc, err := grpc.NewClient(loadedConfig.Host(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			ctx = WithSystemAuthorization(ctx)
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	if err != nil {
		panic(err)
	}
	return system.NewSystemServiceClient(cc)
}

func createSystemUserToken() string {
	const ISSUER = "tester"
	audience := http_util.BuildOrigin(loadedConfig.Host(), loadedConfig.Secure)
	key, err := systemUserKey()
	if err != nil {
		panic(err)
	}
	signer, err := client.NewSignerFromPrivateKeyByte(key, "")
	if err != nil {
		panic(err)
	}
	token, err := client.SignedJWTProfileAssertion(ISSUER, []string{audience}, time.Hour, signer)
	if err != nil {
		panic(err)
	}
	return token
}

func createSystemUserWithNoPermissionsToken() string {
	const ISSUER = "system-user-with-no-permissions"
	audience := http_util.BuildOrigin(loadedConfig.Host(), loadedConfig.Secure)
	key, err := systemUserWithNoPermissions()
	if err != nil {
		panic(err)
	}
	signer, err := client.NewSignerFromPrivateKeyByte(key, "")
	if err != nil {
		panic(err)
	}
	token, err := client.SignedJWTProfileAssertion(ISSUER, []string{audience}, time.Hour, signer)
	if err != nil {
		panic(err)
	}
	return token
}

func integrationPrivateKey(environmentName string) ([]byte, error) {
	if path := os.Getenv(environmentName); path != "" {
		key, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read integration key from %s: %w", environmentName, err)
		}
		return key, nil
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate integration signing fixture: %w", err)
	}
	encoded, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("encode integration signing fixture: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: encoded}), nil
}

func WithSystemAuthorization(ctx context.Context) context.Context {
	return WithAuthorizationToken(ctx, SystemToken)
}

func WithSystemUserWithNoPermissionsAuthorization(ctx context.Context) context.Context {
	return WithAuthorizationToken(ctx, SystemUserWithNoPermissionsToken)
}
