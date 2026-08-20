package ldapoutbound

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/EonsofStupid/tessera/backend/v1/domain"
)

type atomicMemoryTarget struct {
	mu        sync.Mutex
	revision  uint64
	users     map[string]domain.LDAPManagedUser
	applied   map[string]domain.LDAPLifecycleApplyResult
	failAfter int
}

func newAtomicMemoryTarget(users ...domain.LDAPManagedUser) *atomicMemoryTarget {
	target := &atomicMemoryTarget{revision: 1, users: make(map[string]domain.LDAPManagedUser), applied: make(map[string]domain.LDAPLifecycleApplyResult)}
	for _, user := range users {
		target.users[user.TargetUserID] = user
	}
	return target
}

func (t *atomicMemoryTarget) Snapshot(_ context.Context, accountID, workspaceID, _ string) (domain.LDAPTargetSnapshot, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	users := make([]domain.LDAPManagedUser, 0)
	for _, user := range t.users {
		if user.AccountID == accountID && user.WorkspaceID == workspaceID {
			users = append(users, user)
		}
	}
	return domain.LDAPTargetSnapshot{Revision: fmt.Sprintf("r%d", t.revision), Users: users}, nil
}

func (t *atomicMemoryTarget) Apply(_ context.Context, plan domain.LDAPLifecyclePlan) (domain.LDAPLifecycleApplyResult, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if result, exists := t.applied[plan.PlanHash]; exists {
		result.Replayed = true
		return result, nil
	}
	if plan.ExpectedTargetRevision != fmt.Sprintf("r%d", t.revision) {
		return domain.LDAPLifecycleApplyResult{}, &domain.LDAPOutboundError{Reason: domain.LDAPRefusalStaleTarget, Field: "expected_target_revision", Detail: "the lifecycle target changed after preview"}
	}
	working := make(map[string]domain.LDAPManagedUser, len(t.users))
	for id, user := range t.users {
		working[id] = user
	}
	result := domain.LDAPLifecycleApplyResult{}
	for index, action := range plan.Actions {
		if t.failAfter > 0 && index+1 == t.failAfter {
			return domain.LDAPLifecycleApplyResult{}, errors.New("provider detail that must not cross the boundary")
		}
		switch action.Type {
		case domain.LDAPLifecycleCreate:
			id := "managed-" + action.User.Subject
			if _, exists := working[id]; exists {
				return domain.LDAPLifecycleApplyResult{}, &domain.LDAPOutboundError{Reason: domain.LDAPRefusalSyncConflict, Field: "subject", Detail: "the lifecycle target already owns this subject"}
			}
			working[id] = domain.LDAPManagedUser{TargetUserID: id, AccountID: plan.AccountID, WorkspaceID: plan.WorkspaceID, ConnectorID: plan.ConnectorID, User: action.User, Suspended: action.User.Disabled}
			result.Created++
		case domain.LDAPLifecycleUpdate:
			user, exists := working[action.TargetUserID]
			if !exists {
				return domain.LDAPLifecycleApplyResult{}, &domain.LDAPOutboundError{Reason: domain.LDAPRefusalStaleTarget, Field: "target_user_id", Detail: "the lifecycle target user no longer exists"}
			}
			user.User = action.User
			working[action.TargetUserID] = user
			result.Updated++
		case domain.LDAPLifecycleReactivate:
			user := working[action.TargetUserID]
			user.Suspended = false
			working[action.TargetUserID] = user
			result.Reactivated++
		case domain.LDAPLifecycleSuspend:
			user := working[action.TargetUserID]
			user.Suspended = true
			working[action.TargetUserID] = user
			result.Suspended++
		}
	}
	t.users = working
	if len(plan.Actions) > 0 {
		t.revision++
	}
	result.Revision = fmt.Sprintf("r%d", t.revision)
	t.applied[plan.PlanHash] = result
	return result, nil
}

func TestApplyLifecycleIsAtomicRevisionBoundAndIdempotent(t *testing.T) {
	t.Parallel()
	connector := lifecycleTestConnector()
	now := time.Date(2026, 8, 20, 5, 0, 0, 0, time.UTC)
	target := newAtomicMemoryTarget()
	plan, err := domain.BuildLDAPLifecyclePlan(connector, domain.LDAPTargetSnapshot{Revision: "r1"}, []domain.LDAPOutboundUser{{Subject: "alice", Username: "alice"}, {Subject: "bob", Username: "bob"}}, now)
	if err != nil {
		t.Fatal(err)
	}
	client := New(nil)
	result, err := client.ApplyLifecycle(context.Background(), connector, target, plan, now.Add(time.Minute))
	if err != nil || result.Created != 2 || result.Revision != "r2" {
		t.Fatalf("apply = %#v, %v", result, err)
	}
	replay, err := client.ApplyLifecycle(context.Background(), connector, target, plan, now.Add(2*time.Minute))
	if err != nil || !replay.Replayed || replay.Revision != "r2" {
		t.Fatalf("replay = %#v, %v", replay, err)
	}

	stale, err := domain.BuildLDAPLifecyclePlan(connector, domain.LDAPTargetSnapshot{Revision: "r1"}, []domain.LDAPOutboundUser{{Subject: "carol", Username: "carol"}}, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.ApplyLifecycle(context.Background(), connector, target, stale, now.Add(time.Minute))
	assertLifecycleReason(t, err, domain.LDAPRefusalStaleTarget)
}

func TestApplyLifecycleRollsBackAndSanitizesTargetFailure(t *testing.T) {
	t.Parallel()
	connector := lifecycleTestConnector()
	now := time.Date(2026, 8, 20, 5, 0, 0, 0, time.UTC)
	target := newAtomicMemoryTarget()
	target.failAfter = 2
	plan, err := domain.BuildLDAPLifecyclePlan(connector, domain.LDAPTargetSnapshot{Revision: "r1"}, []domain.LDAPOutboundUser{{Subject: "alice", Username: "alice"}, {Subject: "bob", Username: "bob"}}, now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = New(nil).ApplyLifecycle(context.Background(), connector, target, plan, now.Add(time.Minute))
	assertLifecycleReason(t, err, domain.LDAPRefusalUnavailable)
	snapshot, err := target.Snapshot(context.Background(), connector.AccountID, connector.WorkspaceID, connector.ID)
	if err != nil || snapshot.Revision != "r1" || len(snapshot.Users) != 0 {
		t.Fatalf("target changed after failed atomic apply: %#v, %v", snapshot, err)
	}
}

func lifecycleTestConnector() domain.LDAPOutboundConnector {
	return domain.LDAPOutboundConnector{
		ID: "ldap-01", AccountID: "account-01", WorkspaceID: "workspace-01", ResourceRevision: "revision-01", Name: "Lifecycle test", Profile: domain.LDAPProfileOpenLDAP,
		Endpoints: []string{"ldaps://ldap.example.test:636"}, BindDN: "cn=reader,dc=example,dc=test", BindSecretReferenceID: "secret-ref-01", UserBaseDN: "ou=people,dc=example,dc=test",
		UserObjectClasses: []string{"inetOrgPerson"}, LoginAttributes: []string{"uid"}, Attributes: domain.LDAPAttributeMapping{Subject: "entryUUID", Username: "uid"}, Effects: domain.LDAPOutboundEffects{Import: true, Reconcile: true, Deprovision: true},
		ConnectTimeout: time.Second, SearchTimeout: time.Second, ResultLimit: 100, LifecyclePageSize: 100, MaxSyncUsers: 1000,
	}
}

func assertLifecycleReason(t *testing.T, err error, want domain.LDAPOutboundRefusal) {
	t.Helper()
	var refusal *domain.LDAPOutboundError
	if !errors.As(err, &refusal) || refusal.Reason != want {
		t.Fatalf("error = %T %v, want %s", err, err, want)
	}
}
