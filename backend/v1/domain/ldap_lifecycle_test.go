package domain

import (
	"errors"
	"slices"
	"testing"
	"time"
)

func lifecycleConnector() LDAPOutboundConnector {
	connector := validLDAPOutboundConnector()
	connector.Effects = LDAPOutboundEffects{Import: true, Reconcile: true, Deprovision: true}
	connector.LifecyclePageSize = 100
	connector.MaxSyncUsers = 1000
	return connector
}

func directoryUser(subject, username string) LDAPOutboundUser {
	return LDAPOutboundUser{DN: "uid=" + username + ",ou=people,dc=example,dc=test", Subject: subject, Username: username, Email: username + "@example.test", Groups: []string{"developers"}}
}

func TestBuildLDAPLifecyclePlanIsDeterministicAndEffectBound(t *testing.T) {
	t.Parallel()
	connector := lifecycleConnector()
	now := time.Date(2026, 8, 20, 5, 0, 0, 0, time.UTC)
	alice := directoryUser("subject-alice", "alice")
	alice.Email = "alice-new@example.test"
	bob := directoryUser("subject-bob", "bob")
	bob.Disabled = true
	carol := directoryUser("subject-carol", "carol")
	snapshot := LDAPTargetSnapshot{Revision: "target-r1", Users: []LDAPManagedUser{
		{TargetUserID: "user-alice", AccountID: connector.AccountID, WorkspaceID: connector.WorkspaceID, ConnectorID: connector.ID, User: directoryUser("subject-alice", "alice"), Suspended: true},
		{TargetUserID: "user-bob", AccountID: connector.AccountID, WorkspaceID: connector.WorkspaceID, ConnectorID: connector.ID, User: directoryUser("subject-bob", "bob")},
		{TargetUserID: "user-dave", AccountID: connector.AccountID, WorkspaceID: connector.WorkspaceID, ConnectorID: connector.ID, User: directoryUser("subject-dave", "dave")},
		{TargetUserID: "local-owner", AccountID: connector.AccountID, WorkspaceID: connector.WorkspaceID, ConnectorID: "", User: directoryUser("local", "owner")},
	}}
	plan, err := BuildLDAPLifecyclePlan(connector, snapshot, []LDAPOutboundUser{carol, bob, alice}, now)
	if err != nil {
		t.Fatal(err)
	}
	want := []LDAPLifecycleActionType{LDAPLifecycleUpdate, LDAPLifecycleReactivate, LDAPLifecycleUpdate, LDAPLifecycleSuspend, LDAPLifecycleCreate, LDAPLifecycleSuspend}
	got := make([]LDAPLifecycleActionType, len(plan.Actions))
	for index := range plan.Actions {
		got[index] = plan.Actions[index].Type
	}
	if !slices.Equal(got, want) {
		t.Fatalf("actions = %v, want %v", got, want)
	}
	if plan.ExpectedTargetRevision != "target-r1" || plan.ExpiresAt.Sub(plan.GeneratedAt) != LDAPLifecyclePlanTTL {
		t.Fatalf("plan bounds = %#v", plan)
	}
	if err := plan.Validate(connector, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	slices.Reverse(snapshot.Users)
	source := []LDAPOutboundUser{alice, bob, carol}
	second, err := BuildLDAPLifecyclePlan(connector, snapshot, source, now)
	if err != nil {
		t.Fatal(err)
	}
	if second.PlanHash != plan.PlanHash {
		t.Fatalf("plan hash changed with input order: %s != %s", second.PlanHash, plan.PlanHash)
	}
	for _, action := range plan.Actions {
		if action.TargetUserID == "local-owner" {
			t.Fatal("local owner entered lifecycle plan")
		}
	}
}

func TestLDAPLifecyclePlanRejectsConflictsAndTampering(t *testing.T) {
	t.Parallel()
	connector := lifecycleConnector()
	now := time.Date(2026, 8, 20, 5, 0, 0, 0, time.UTC)
	assertReason := func(err error, want LDAPOutboundRefusal) {
		t.Helper()
		var refusal *LDAPOutboundError
		if !errors.As(err, &refusal) || refusal.Reason != want {
			t.Fatalf("error = %T %v, want %s", err, err, want)
		}
	}

	_, err := BuildLDAPLifecyclePlan(connector, LDAPTargetSnapshot{}, []LDAPOutboundUser{directoryUser("same", "alice"), directoryUser("same", "bob")}, now)
	assertReason(err, LDAPRefusalSyncConflict)
	_, err = BuildLDAPLifecyclePlan(connector, LDAPTargetSnapshot{}, []LDAPOutboundUser{directoryUser("one", "Alice"), directoryUser("two", "alice")}, now)
	assertReason(err, LDAPRefusalSyncConflict)
	_, err = BuildLDAPLifecyclePlan(connector, LDAPTargetSnapshot{Users: []LDAPManagedUser{{AccountID: "other", WorkspaceID: connector.WorkspaceID}}}, nil, now)
	assertReason(err, LDAPRefusalSyncConflict)
	_, err = BuildLDAPLifecyclePlan(connector, LDAPTargetSnapshot{Users: []LDAPManagedUser{{TargetUserID: "local-alice", AccountID: connector.AccountID, WorkspaceID: connector.WorkspaceID, User: directoryUser("local-subject", "alice")}}}, []LDAPOutboundUser{directoryUser("directory-subject", "Alice")}, now)
	assertReason(err, LDAPRefusalSyncConflict)

	plan, err := BuildLDAPLifecyclePlan(connector, LDAPTargetSnapshot{Revision: "r1"}, []LDAPOutboundUser{directoryUser("one", "alice")}, now)
	if err != nil {
		t.Fatal(err)
	}
	tampered := plan
	tampered.Actions = append([]LDAPLifecycleAction(nil), plan.Actions...)
	tampered.Actions[0].User.Email = "attacker@example.test"
	assertReason(tampered.Validate(connector, now.Add(time.Minute)), LDAPRefusalInvalidPlan)
	assertReason(plan.Validate(connector, plan.ExpiresAt), LDAPRefusalInvalidPlan)
	changed := connector
	changed.ResourceRevision = "revision-02"
	assertReason(plan.Validate(changed, now.Add(time.Minute)), LDAPRefusalInvalidPlan)
}

func TestLDAPLifecycleImportOnlyDoesNotReconcileOrSuspend(t *testing.T) {
	t.Parallel()
	connector := lifecycleConnector()
	connector.Effects = LDAPOutboundEffects{Import: true}
	now := time.Date(2026, 8, 20, 5, 0, 0, 0, time.UTC)
	snapshot := LDAPTargetSnapshot{Revision: "r1", Users: []LDAPManagedUser{
		{TargetUserID: "alice", AccountID: connector.AccountID, WorkspaceID: connector.WorkspaceID, ConnectorID: connector.ID, User: directoryUser("one", "alice")},
		{TargetUserID: "missing", AccountID: connector.AccountID, WorkspaceID: connector.WorkspaceID, ConnectorID: connector.ID, User: directoryUser("missing", "missing")},
	}}
	changed := directoryUser("one", "alice")
	changed.Email = "changed@example.test"
	plan, err := BuildLDAPLifecyclePlan(connector, snapshot, []LDAPOutboundUser{changed, directoryUser("two", "bob")}, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Actions) != 1 || plan.Actions[0].Type != LDAPLifecycleCreate || plan.Actions[0].User.Subject != "two" {
		t.Fatalf("import-only actions = %#v", plan.Actions)
	}
}
