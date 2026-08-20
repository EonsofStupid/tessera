package fake

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/EonsofStupid/tessera/backend/v1/domain"
)

var fakeNow = time.Date(2026, 8, 20, 6, 0, 0, 0, time.UTC)

func enrollment() domain.EnrollSecretRequest {
	return domain.EnrollSecretRequest{
		ReferenceID: "ldap-bind-01",
		AccountID:   "account-01",
		WorkspaceID: "workspace-01",
		Purpose:     domain.SecretPurposeLDAPBind,
		OperationID: "operation-enroll-01",
	}
}

func useRequest() domain.UseSecretRequest {
	request := enrollment()
	return domain.UseSecretRequest{
		ReferenceID: request.ReferenceID,
		AccountID:   request.AccountID,
		WorkspaceID: request.WorkspaceID,
		Purpose:     request.Purpose,
		OperationID: "operation-use-01",
	}
}

func reason(t *testing.T, err error) domain.SecretCustodyRefusal {
	t.Helper()
	var refusal *domain.SecretCustodyError
	if !errors.As(err, &refusal) {
		t.Fatalf("error = %T %v, want custody refusal", err, err)
	}
	return refusal.Reason
}

func TestMemoryEnrollAndUseAreValueBlind(t *testing.T) {
	seed := "seeded-secret-that-must-never-leak"
	store := New(func() time.Time { return fakeNow })
	reference, err := store.Enroll(context.Background(), enrollment(), strings.NewReader(seed))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(reference)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte(seed)) {
		t.Fatal("safe reference leaked enrolled material")
	}

	var callbackAlias []byte
	receipt, err := store.Use(context.Background(), useRequest(), func(value []byte) error {
		if string(value) != seed {
			t.Fatalf("callback received %q", value)
		}
		callbackAlias = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(callbackAlias, make([]byte, len(callbackAlias))) {
		t.Fatal("callback-scoped working copy was not cleared on return")
	}
	receiptJSON, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(receiptJSON, []byte(seed)) {
		t.Fatal("safe receipt leaked enrolled material")
	}
}

func TestMemoryNeverPropagatesConsumerErrorText(t *testing.T) {
	seed := "seeded-secret-in-dependency-error"
	store := New(func() time.Time { return fakeNow })
	if _, err := store.Enroll(context.Background(), enrollment(), strings.NewReader(seed)); err != nil {
		t.Fatal(err)
	}
	_, err := store.Use(context.Background(), useRequest(), func([]byte) error { return errors.New(seed) })
	if got := reason(t, err); got != domain.SecretRefusalCallbackFailed {
		t.Fatalf("reason = %s", got)
	}
	if strings.Contains(err.Error(), seed) {
		t.Fatal("consumer error text leaked through custody boundary")
	}
}

func TestMemoryRefusesTenantPurposeAndLifecycleMismatches(t *testing.T) {
	store := New(func() time.Time { return fakeNow })
	if _, err := store.Enroll(context.Background(), enrollment(), strings.NewReader("protected")); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		mutate func(*domain.UseSecretRequest)
		want   domain.SecretCustodyRefusal
	}{
		{"account", func(value *domain.UseSecretRequest) { value.AccountID = "account-02" }, domain.SecretRefusalTenantMismatch},
		{"workspace", func(value *domain.UseSecretRequest) { value.WorkspaceID = "workspace-02" }, domain.SecretRefusalTenantMismatch},
		{"purpose", func(value *domain.UseSecretRequest) { value.Purpose = domain.SecretPurposeEdgeTLS }, domain.SecretRefusalPurposeMismatch},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := useRequest()
			test.mutate(&request)
			_, err := store.Use(context.Background(), request, func([]byte) error { return nil })
			if got := reason(t, err); got != test.want {
				t.Fatalf("reason = %s, want %s", got, test.want)
			}
		})
	}

	for _, test := range []struct {
		state domain.SecretCustodyState
		want  domain.SecretCustodyRefusal
	}{
		{domain.SecretCustodyExpired, domain.SecretRefusalExpired},
		{domain.SecretCustodyRevoked, domain.SecretRefusalRevoked},
		{domain.SecretCustodyUnavailable, domain.SecretRefusalUnavailable},
	} {
		if err := store.SetState(enrollment().ReferenceID, test.state); err != nil {
			t.Fatal(err)
		}
		_, err := store.Use(context.Background(), useRequest(), func([]byte) error { return nil })
		if got := reason(t, err); got != test.want {
			t.Fatalf("state %s reason = %s, want %s", test.state, got, test.want)
		}
	}
}

func TestMemoryBoundsEnrollmentAndRejectsDuplicates(t *testing.T) {
	store := New(func() time.Time { return fakeNow })
	if _, err := store.Enroll(context.Background(), enrollment(), bytes.NewReader(nil)); reason(t, err) != domain.SecretRefusalInvalidInput {
		t.Fatal("empty material was accepted")
	}
	if _, err := store.Enroll(context.Background(), enrollment(), bytes.NewReader(make([]byte, maxSecretBytes+1))); reason(t, err) != domain.SecretRefusalInvalidInput {
		t.Fatal("oversized material was accepted")
	}
	if _, err := store.Enroll(context.Background(), enrollment(), strings.NewReader("protected")); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Enroll(context.Background(), enrollment(), strings.NewReader("other")); reason(t, err) != domain.SecretRefusalDenied {
		t.Fatal("duplicate reference was accepted")
	}
}

func TestMemoryHonorsCancellationAndClose(t *testing.T) {
	store := New(func() time.Time { return fakeNow })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.Enroll(ctx, enrollment(), strings.NewReader("protected")); reason(t, err) != domain.SecretRefusalUnavailable {
		t.Fatal("canceled enrollment was accepted")
	}
	if _, err := store.Enroll(context.Background(), enrollment(), strings.NewReader("protected")); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Use(ctx, useRequest(), func([]byte) error { return nil }); reason(t, err) != domain.SecretRefusalUnavailable {
		t.Fatal("canceled use was accepted")
	}
	store.Close()
	if _, err := store.Get(context.Background(), enrollment().ReferenceID); reason(t, err) != domain.SecretRefusalUnknown {
		t.Fatal("closed fake retained its reference")
	}
}
