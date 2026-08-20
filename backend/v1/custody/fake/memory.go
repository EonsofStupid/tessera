// Package fake supplies an in-memory custody adapter for contract tests. It is
// never production persistence and deliberately exposes only the same
// value-blind port a real Vaultix adapter implements.
package fake

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/EonsofStupid/tessera/backend/v1/domain"
)

const maxSecretBytes = 64 * 1024

type record struct {
	reference domain.SecretReference
	value     []byte
}

type Memory struct {
	mu      sync.RWMutex
	now     func() time.Time
	records map[string]record
}

func New(now func() time.Time) *Memory {
	if now == nil {
		now = time.Now
	}
	return &Memory{now: now, records: make(map[string]record)}
}

func (m *Memory) Enroll(ctx context.Context, request domain.EnrollSecretRequest, input io.Reader) (domain.SecretReference, error) {
	if err := ctx.Err(); err != nil {
		return domain.SecretReference{}, custodyError(domain.SecretRefusalUnavailable, "context", "the protected enrollment was canceled")
	}
	if err := domain.ValidateSecretEnrollment(request); err != nil {
		return domain.SecretReference{}, err
	}
	if input == nil {
		return domain.SecretReference{}, custodyError(domain.SecretRefusalInvalidInput, "value", "protected material is required")
	}
	value, err := io.ReadAll(io.LimitReader(input, maxSecretBytes+1))
	if err != nil {
		return domain.SecretReference{}, custodyError(domain.SecretRefusalUnavailable, "value", "protected material could not be enrolled")
	}
	if len(value) == 0 || len(value) > maxSecretBytes {
		clear(value)
		return domain.SecretReference{}, custodyError(domain.SecretRefusalInvalidInput, "value", "protected material must contain 1 to 65536 bytes")
	}

	now := m.now().UTC()
	reference := domain.SecretReference{
		ID:                request.ReferenceID,
		AccountID:         request.AccountID,
		WorkspaceID:       request.WorkspaceID,
		Purpose:           request.Purpose,
		ProviderReference: "vaultix://sandbox/secrets/" + request.ReferenceID,
		State:             domain.SecretCustodyActive,
		ProviderVersion:   "fake-v1",
		ResourceRevision:  "revision-1",
		CustodyAuditID:    "audit-enroll-" + request.OperationID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := reference.Validate(now); err != nil {
		clear(value)
		return domain.SecretReference{}, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.records[reference.ID]; exists {
		clear(value)
		return domain.SecretReference{}, custodyError(domain.SecretRefusalDenied, "reference_id", "the protected reference already exists")
	}
	m.records[reference.ID] = record{reference: reference, value: value}
	return reference, nil
}

func (m *Memory) Use(ctx context.Context, request domain.UseSecretRequest, consume func([]byte) error) (domain.SecretUseReceipt, error) {
	if err := ctx.Err(); err != nil {
		return domain.SecretUseReceipt{}, custodyError(domain.SecretRefusalUnavailable, "context", "the protected operation was canceled")
	}
	m.mu.RLock()
	stored, found := m.records[request.ReferenceID]
	if !found {
		m.mu.RUnlock()
		return domain.SecretUseReceipt{}, custodyError(domain.SecretRefusalUnknown, "reference_id", "the protected reference was not found")
	}
	reference := stored.reference
	working := append([]byte(nil), stored.value...)
	m.mu.RUnlock()
	defer clear(working)

	if err := domain.ValidateSecretUse(reference, request, m.now().UTC()); err != nil {
		return domain.SecretUseReceipt{}, err
	}
	if consume == nil {
		return domain.SecretUseReceipt{}, custodyError(domain.SecretRefusalDenied, "consumer", "an authorized secret consumer is required")
	}
	if err := ctx.Err(); err != nil {
		return domain.SecretUseReceipt{}, custodyError(domain.SecretRefusalUnavailable, "context", "the protected operation was canceled")
	}
	if err := consume(working); err != nil {
		// Never propagate dependency text: it may contain the credential it was
		// trying to use. The operation owns the safe diagnostic correlation.
		return domain.SecretUseReceipt{}, custodyError(domain.SecretRefusalCallbackFailed, "consumer", "the protected operation failed")
	}
	now := m.now().UTC()
	return domain.SecretUseReceipt{
		ReferenceID:    reference.ID,
		OperationID:    request.OperationID,
		CustodyAuditID: "audit-use-" + request.OperationID,
		UsedAt:         now,
	}, nil
}

// Close clears the fake adapter's retained test material. Production Vaultix
// owns persistence; this exists so tests do not leave protected bytes in a
// long-lived process after the adapter is no longer needed.
func (m *Memory) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, stored := range m.records {
		clear(stored.value)
		delete(m.records, id)
	}
}

func (m *Memory) Get(_ context.Context, referenceID string) (domain.SecretReference, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	record, found := m.records[referenceID]
	if !found {
		return domain.SecretReference{}, custodyError(domain.SecretRefusalUnknown, "reference_id", "the protected reference was not found")
	}
	return record.reference, nil
}

// SetState drives expiry, revocation and outage contract tests without adding
// a production mutation method to the custody port.
func (m *Memory) SetState(referenceID string, state domain.SecretCustodyState) error {
	if !state.Valid() {
		return custodyError(domain.SecretRefusalInvalidInput, "state", "a known custody state is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	stored, found := m.records[referenceID]
	if !found {
		return custodyError(domain.SecretRefusalUnknown, "reference_id", "the protected reference was not found")
	}
	stored.reference.State = state
	stored.reference.UpdatedAt = m.now().UTC()
	stored.reference.ResourceRevision = "revision-state-" + string(state)
	m.records[referenceID] = stored
	return nil
}

func custodyError(reason domain.SecretCustodyRefusal, field, detail string) error {
	return &domain.SecretCustodyError{Reason: reason, Field: field, Detail: detail}
}

var _ domain.SecretCustody = (*Memory)(nil)
