package management

import (
	"net/http"

	"github.com/EonsofStupid/tessera/backend/v1/domain"
	"github.com/EonsofStupid/tessera/internal/api/authz"
	http_util "github.com/EonsofStupid/tessera/internal/api/http"
	"github.com/EonsofStupid/tessera/internal/zerrors"
)

type TesseraAuthorizer struct {
	verifier authz.APITokenVerifier
	system   authz.Config
	internal authz.Config
}

func NewTesseraAuthorizer(verifier authz.APITokenVerifier, system, internal authz.Config) *TesseraAuthorizer {
	return &TesseraAuthorizer{verifier: verifier, system: system, internal: internal}
}

type authorizationRequest struct{}

func (a *TesseraAuthorizer) Authorize(r *http.Request, permission string) (*http.Request, *domain.ManagementError) {
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
