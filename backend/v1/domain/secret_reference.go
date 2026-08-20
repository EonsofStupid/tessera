package domain

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
)

// SecretPurpose binds protected material to one use. A connector cannot use a
// reference enrolled for a different protocol merely because both are bytes.
type SecretPurpose string

const (
	SecretPurposeLDAPBind        SecretPurpose = "ldap_bind"
	SecretPurposeEdgeTLS         SecretPurpose = "edge_tls"
	SecretPurposeProxySession    SecretPurpose = "proxy_session"
	SecretPurposeOIDCClient      SecretPurpose = "oidc_client"
	SecretPurposeSAMLSigning     SecretPurpose = "saml_signing"
	SecretPurposePrivilegedLease SecretPurpose = "privileged_lease"
)

func (p SecretPurpose) Valid() bool {
	switch p {
	case SecretPurposeLDAPBind, SecretPurposeEdgeTLS, SecretPurposeProxySession, SecretPurposeOIDCClient, SecretPurposeSAMLSigning, SecretPurposePrivilegedLease:
		return true
	default:
		return false
	}
}

type SecretCustodyState string

const (
	SecretCustodyPending     SecretCustodyState = "pending"
	SecretCustodyActive      SecretCustodyState = "active"
	SecretCustodyRotationDue SecretCustodyState = "rotation_due"
	SecretCustodyExpired     SecretCustodyState = "expired"
	SecretCustodyRevoked     SecretCustodyState = "revoked"
	SecretCustodyUnavailable SecretCustodyState = "unavailable"
)

func (s SecretCustodyState) Valid() bool {
	switch s {
	case SecretCustodyPending, SecretCustodyActive, SecretCustodyRotationDue, SecretCustodyExpired, SecretCustodyRevoked, SecretCustodyUnavailable:
		return true
	default:
		return false
	}
}

// SecretReference is the complete safe projection Tessera may persist or
// return. It deliberately has no generic metadata or value-bearing field.
type SecretReference struct {
	ID                string             `json:"reference_id"`
	AccountID         string             `json:"account_id"`
	WorkspaceID       string             `json:"workspace_id,omitempty"`
	Purpose           SecretPurpose      `json:"purpose"`
	ProviderReference string             `json:"provider_reference"`
	State             SecretCustodyState `json:"state"`
	ProviderVersion   string             `json:"provider_version,omitempty"`
	ResourceRevision  string             `json:"resource_revision"`
	RotatesAt         *time.Time         `json:"rotates_at,omitempty"`
	ExpiresAt         *time.Time         `json:"expires_at,omitempty"`
	CustodyAuditID    string             `json:"custody_audit_id,omitempty"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

func (r SecretReference) Validate(now time.Time) error {
	switch {
	case !safeIdentifier(r.ID):
		return secretError(SecretRefusalInvalidReference, "reference_id", "a safe stable reference id is required")
	case !safeIdentifier(r.AccountID):
		return secretError(SecretRefusalInvalidReference, "account_id", "a safe account boundary is required")
	case r.WorkspaceID != "" && !safeIdentifier(r.WorkspaceID):
		return secretError(SecretRefusalInvalidReference, "workspace_id", "workspace scope is invalid")
	case !r.Purpose.Valid():
		return secretError(SecretRefusalInvalidReference, "purpose", "a known secret purpose is required")
	case !r.State.Valid():
		return secretError(SecretRefusalInvalidReference, "state", "a known custody state is required")
	case !validVaultixReference(r.ProviderReference):
		return secretError(SecretRefusalInvalidReference, "provider_reference", "want an opaque vaultix reference without user-info, query or fragment")
	case !safeIdentifier(r.ResourceRevision):
		return secretError(SecretRefusalInvalidReference, "resource_revision", "a safe resource revision is required")
	case r.CreatedAt.IsZero() || r.UpdatedAt.IsZero() || r.UpdatedAt.Before(r.CreatedAt):
		return secretError(SecretRefusalInvalidReference, "updated_at", "valid creation and update times are required")
	case r.RotatesAt != nil && !r.RotatesAt.After(r.CreatedAt):
		return secretError(SecretRefusalInvalidReference, "rotates_at", "rotation must follow creation")
	case r.ExpiresAt != nil && !r.ExpiresAt.After(r.CreatedAt):
		return secretError(SecretRefusalInvalidReference, "expires_at", "expiry must follow creation")
	case r.ExpiresAt != nil && !now.Before(*r.ExpiresAt) && r.State != SecretCustodyExpired && r.State != SecretCustodyRevoked:
		return secretError(SecretRefusalExpired, "expires_at", "the reference has passed its expiry")
	default:
		return nil
	}
}

func (r SecretReference) Usable(now time.Time) error {
	if err := r.Validate(now); err != nil {
		return err
	}
	switch r.State {
	case SecretCustodyActive, SecretCustodyRotationDue:
		return nil
	case SecretCustodyExpired:
		return secretError(SecretRefusalExpired, "state", "the protected material is expired")
	case SecretCustodyRevoked:
		return secretError(SecretRefusalRevoked, "state", "the protected material is revoked")
	case SecretCustodyUnavailable:
		return secretError(SecretRefusalUnavailable, "state", "the custody provider is unavailable")
	default:
		return secretError(SecretRefusalUnavailable, "state", "the protected material is not active")
	}
}

type SecretCustodyRefusal string

const (
	SecretRefusalInvalidReference SecretCustodyRefusal = "invalid_reference"
	SecretRefusalInvalidInput     SecretCustodyRefusal = "invalid_input"
	SecretRefusalPurposeMismatch  SecretCustodyRefusal = "purpose_mismatch"
	SecretRefusalTenantMismatch   SecretCustodyRefusal = "tenant_mismatch"
	SecretRefusalUnknown          SecretCustodyRefusal = "unknown_reference"
	SecretRefusalDenied           SecretCustodyRefusal = "denied"
	SecretRefusalExpired          SecretCustodyRefusal = "expired"
	SecretRefusalRevoked          SecretCustodyRefusal = "revoked"
	SecretRefusalUnavailable      SecretCustodyRefusal = "unavailable"
	SecretRefusalCallbackFailed   SecretCustodyRefusal = "consumer_failed"
)

type SecretCustodyError struct {
	Reason SecretCustodyRefusal `json:"reason"`
	Field  string               `json:"field,omitempty"`
	Detail string               `json:"detail"`
}

func (e *SecretCustodyError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("secret custody: %s (%s): %s", e.Reason, e.Field, e.Detail)
	}
	return fmt.Sprintf("secret custody: %s: %s", e.Reason, e.Detail)
}

type EnrollSecretRequest struct {
	ReferenceID string
	AccountID   string
	WorkspaceID string
	Purpose     SecretPurpose
	OperationID string
}

type UseSecretRequest struct {
	ReferenceID string
	AccountID   string
	WorkspaceID string
	Purpose     SecretPurpose
	OperationID string
}

type SecretUseReceipt struct {
	ReferenceID    string    `json:"reference_id"`
	OperationID    string    `json:"operation_id"`
	CustodyAuditID string    `json:"custody_audit_id"`
	UsedAt         time.Time `json:"used_at"`
}

// SecretCustody never returns protected bytes. Use scopes them to the callback
// and returns only a safe receipt after the working copy has been cleared.
type SecretCustody interface {
	Enroll(context.Context, EnrollSecretRequest, io.Reader) (SecretReference, error)
	Use(context.Context, UseSecretRequest, func([]byte) error) (SecretUseReceipt, error)
	Get(context.Context, string) (SecretReference, error)
}

// CustodySafeError marks a typed Tessera refusal whose fields have already
// been constrained to safe vocabulary. Adapters may preserve this type across
// a Use callback; every unmarked error must be replaced rather than wrapped.
type CustodySafeError interface {
	error
	custodySafe()
}

func ValidateSecretEnrollment(request EnrollSecretRequest) error {
	switch {
	case !safeIdentifier(request.ReferenceID):
		return secretError(SecretRefusalInvalidInput, "reference_id", "a safe stable reference id is required")
	case !safeIdentifier(request.AccountID):
		return secretError(SecretRefusalInvalidInput, "account_id", "a safe account boundary is required")
	case request.WorkspaceID != "" && !safeIdentifier(request.WorkspaceID):
		return secretError(SecretRefusalInvalidInput, "workspace_id", "workspace scope is invalid")
	case !request.Purpose.Valid():
		return secretError(SecretRefusalInvalidInput, "purpose", "a known secret purpose is required")
	case !safeIdentifier(request.OperationID):
		return secretError(SecretRefusalDenied, "operation_id", "an authorized operation id is required")
	default:
		return nil
	}
}

func ValidateSecretUse(reference SecretReference, request UseSecretRequest, now time.Time) error {
	if err := reference.Usable(now); err != nil {
		return err
	}
	switch {
	case request.ReferenceID != reference.ID:
		return secretError(SecretRefusalUnknown, "reference_id", "the requested reference was not found")
	case request.AccountID != reference.AccountID || request.WorkspaceID != reference.WorkspaceID:
		return secretError(SecretRefusalTenantMismatch, "scope", "the reference belongs to another tenant scope")
	case request.Purpose != reference.Purpose:
		return secretError(SecretRefusalPurposeMismatch, "purpose", "the reference is bound to another purpose")
	case !safeIdentifier(request.OperationID):
		return secretError(SecretRefusalDenied, "operation_id", "an authorized operation id is required")
	default:
		return nil
	}
}

func validVaultixReference(value string) bool {
	if len(value) == 0 || len(value) > 512 || strings.ContainsAny(value, "\r\n\t ") {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "vaultix" || parsed.Host == "" || parsed.Path == "" {
		return false
	}
	return parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == ""
}

func safeIdentifier(value string) bool {
	if value == "" || len(value) > 255 {
		return false
	}
	for _, char := range []byte(value) {
		if char < 0x21 || char > 0x7e || char == '?' || char == '#' {
			return false
		}
	}
	return true
}

func secretError(reason SecretCustodyRefusal, field, detail string) error {
	return &SecretCustodyError{Reason: reason, Field: field, Detail: detail}
}
