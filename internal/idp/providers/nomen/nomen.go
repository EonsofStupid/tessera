package nomen

import (
	"net/http"

	"github.com/shippinAI/nomen/oidc/v3/pkg/client/rp"

	"github.com/shippinAI/nomen/internal/idp"
	"github.com/shippinAI/nomen/internal/idp/providers/oidc"
)

const (
	name = "NOMEN"
)

var _ idp.Provider = (*Provider)(nil)

// Provider is the [idp.Provider] implementation for NOMEN
type Provider struct {
	*oidc.Provider
}

func New(issuer, clientID, clientSecret, redirectURI string, scopes []string, httpClient *http.Client, options ...oidc.ProviderOpts) (*Provider, error) {
	// PKCE is used by default
	options = append(options, oidc.WithRelyingPartyOption(rp.WithPKCE(nil)))
	provider, err := oidc.New(name, issuer, clientID, clientSecret, redirectURI, scopes, oidc.DefaultMapper, httpClient, options...)
	if err != nil {
		return nil, err
	}
	return &Provider{
		Provider: provider,
	}, nil
}
