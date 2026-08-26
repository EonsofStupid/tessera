package management

import (
	"net/http"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/shippinAI/nomen/internal/api/authz"
	http_util "github.com/shippinAI/nomen/internal/api/http"
	"github.com/shippinAI/nomen/internal/zerrors"
)

type NomenAuthorizer struct {
	verifier authz.APITokenVerifier
	system   authz.Config
	internal authz.Config
}

func NewNomenAuthorizer(verifier authz.APITokenVerifier, system, internal authz.Config) *NomenAuthorizer {
	return &NomenAuthorizer{verifier: verifier, system: system, internal: internal}
}

type authorizationRequest struct{}

func (a *NomenAuthorizer) Authorize(r *http.Request, permission string) (*http.Request, *domain.ManagementError) {
	token := http_util.GetAuthorization(r)
	if token == "" || a == nil || a.verifier == nil {
		managementError := authenticationError()
		return nil, &managementError
	}
	setContext, err := authz.CheckUserAuthorization(
		r.Context(),
		&authorizationRequest{},
		token,
		http_util.GetOrgID(r),
		"",
		a.verifier,
		a.system.RolePermissionMappings,
		a.internal.RolePermissionMappings,
		authz.Option{Permission: permission},
		r.RequestURI,
	)
	if err == nil {
		return r.WithContext(setContext(r.Context())), nil
	}
	if zerrors.IsUnauthenticated(err) {
		managementError := authenticationError()
		return nil, &managementError
	}
	if zerrors.IsPermissionDenied(err) {
		managementError := permissionError(permission)
		return nil, &managementError
	}
	managementError := unavailableError("overview-authorization")
	return nil, &managementError
}
