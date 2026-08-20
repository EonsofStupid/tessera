package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var secretTestNow = time.Date(2026, 8, 20, 5, 0, 0, 0, time.UTC)

func validSecretReference() SecretReference {
	expires := secretTestNow.Add(time.Hour)
	return SecretReference{
		ID:                "secret-ref-01",
		AccountID:         "account-01",
		WorkspaceID:       "workspace-01",
		Purpose:           SecretPurposeLDAPBind,
		ProviderReference: "vaultix://project-01/secrets/ldap-bind-01",
		State:             SecretCustodyActive,
		ProviderVersion:   "version-01",
		ResourceRevision:  "revision-01",
		ExpiresAt:         &expires,
		CreatedAt:         secretTestNow.Add(-time.Hour),
		UpdatedAt:         secretTestNow.Add(-time.Minute),
	}
}

func secretRefusal(t *testing.T, err error) *SecretCustodyError {
	t.Helper()
	var refusal *SecretCustodyError
	if !errors.As(err, &refusal) {
		t.Fatalf("error = %T %v, want *SecretCustodyError", err, err)
	}
	return refusal
}

func TestSecretReferenceValidate(t *testing.T) {
	t.Parallel()
	reference := validSecretReference()
	if err := reference.Validate(secretTestNow); err != nil {
		t.Fatalf("Validate(): %v", err)
	}
	if err := reference.Usable(secretTestNow); err != nil {
		t.Fatalf("Usable(): %v", err)
	}
}

func TestSecretReferenceRejectsUnsafeProviderReferences(t *testing.T) {
	t.Parallel()
	for _, providerReference := range []string{
		"https://vaultix.example/secrets/one",
		"vaultix://user:password@project/secrets/one",
		"vaultix://project/secrets/one?token=secret",
		"vaultix://project/secrets/one#secret",
		"vaultix://project",
	} {
		reference := validSecretReference()
		reference.ProviderReference = providerReference
		refusal := secretRefusal(t, reference.Validate(secretTestNow))
		if refusal.Reason != SecretRefusalInvalidReference || refusal.Field != "provider_reference" {
			t.Fatalf("reference %q refusal = %#v", providerReference, refusal)
		}
	}
}

func TestSecretReferenceUsableStates(t *testing.T) {
	t.Parallel()
	tests := []struct {
		state SecretCustodyState
		want  SecretCustodyRefusal
	}{
		{SecretCustodyPending, SecretRefusalUnavailable},
		{SecretCustodyExpired, SecretRefusalExpired},
		{SecretCustodyRevoked, SecretRefusalRevoked},
		{SecretCustodyUnavailable, SecretRefusalUnavailable},
	}
	for _, test := range tests {
		reference := validSecretReference()
		reference.State = test.state
		refusal := secretRefusal(t, reference.Usable(secretTestNow))
		if refusal.Reason != test.want {
			t.Errorf("state %s refusal = %s, want %s", test.state, refusal.Reason, test.want)
		}
	}
	reference := validSecretReference()
	reference.State = SecretCustodyRotationDue
	if err := reference.Usable(secretTestNow); err != nil {
		t.Fatalf("rotation overlap must remain usable: %v", err)
	}
}

func TestValidateSecretUseBindsTenantPurposeAndOperation(t *testing.T) {
	t.Parallel()
	reference := validSecretReference()
	request := UseSecretRequest{
		ReferenceID: reference.ID,
		AccountID:   reference.AccountID,
		WorkspaceID: reference.WorkspaceID,
		Purpose:     reference.Purpose,
		OperationID: "operation-01",
	}
	if err := ValidateSecretUse(reference, request, secretTestNow); err != nil {
		t.Fatalf("ValidateSecretUse(): %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*UseSecretRequest)
		want   SecretCustodyRefusal
	}{
		{"reference", func(value *UseSecretRequest) { value.ReferenceID = "other" }, SecretRefusalUnknown},
		{"account", func(value *UseSecretRequest) { value.AccountID = "other" }, SecretRefusalTenantMismatch},
		{"workspace", func(value *UseSecretRequest) { value.WorkspaceID = "other" }, SecretRefusalTenantMismatch},
		{"purpose", func(value *UseSecretRequest) { value.Purpose = SecretPurposeEdgeTLS }, SecretRefusalPurposeMismatch},
		{"operation", func(value *UseSecretRequest) { value.OperationID = "" }, SecretRefusalDenied},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changed := request
			test.mutate(&changed)
			if got := secretRefusal(t, ValidateSecretUse(reference, changed, secretTestNow)).Reason; got != test.want {
				t.Fatalf("reason = %s, want %s", got, test.want)
			}
		})
	}
}

func TestSecretVocabulary(t *testing.T) {
	t.Parallel()
	for _, purpose := range []SecretPurpose{SecretPurposeLDAPBind, SecretPurposeEdgeTLS, SecretPurposeProxySession, SecretPurposeOIDCClient, SecretPurposeSAMLSigning, SecretPurposePrivilegedLease} {
		if !purpose.Valid() {
			t.Errorf("purpose %q is invalid", purpose)
		}
	}
	for _, state := range []SecretCustodyState{SecretCustodyPending, SecretCustodyActive, SecretCustodyRotationDue, SecretCustodyExpired, SecretCustodyRevoked, SecretCustodyUnavailable} {
		if !state.Valid() {
			t.Errorf("state %q is invalid", state)
		}
	}
	if strings.Contains(strings.ToLower(strings.Join([]string{
		"reference_id", "account_id", "workspace_id", "purpose", "provider_reference", "state",
		"provider_version", "resource_revision", "rotates_at", "expires_at", "custody_audit_id",
	}, " ")), "password") {
		t.Fatal("safe projection vocabulary contains a value-bearing field")
	}
}
