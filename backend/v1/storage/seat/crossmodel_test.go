package seat

// Cross-model tests live here because this package owns the one embedded-PG
// harness (TestMain in atomicity_test.go), and a second embedded instance
// would buy nothing but wall-clock. The import of storage/flow is test-only;
// production code in these packages stays uncoupled.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/EonsofStupid/tessera/backend/v1/domain"
	flowstorage "github.com/EonsofStupid/tessera/backend/v1/storage/flow"
)

func flowEntry(slug string, stages ...map[string]any) domain.Entry {
	list := make([]any, len(stages))
	for i, s := range stages {
		list[i] = s
	}
	return domain.Entry{
		Model:       flowstorage.FlowModel,
		Identifiers: map[string]string{"slug": slug},
		Attrs: map[string]any{
			"title":       "Test flow",
			"designation": "authentication",
			"stages":      list,
		},
	}
}

func TestFlowApplier_ConvergesAndReadsBackInOrder(t *testing.T) {
	eng := domain.NewBlueprintEngine(NewApplier(), flowstorage.NewApplier())
	ctx := context.Background()
	bp := &domain.Blueprint{
		Schema: domain.BlueprintSchema,
		Entries: []domain.Entry{flowEntry("login-mfa",
			map[string]any{"kind": "identify"},
			map[string]any{"kind": "password"},
			map[string]any{"kind": "totp"},
		)},
	}

	report, err := eng.Apply(ctx, testPool, testInstance, bp)
	if err != nil {
		t.Fatal(err)
	}
	if report.Results[0].Outcome != domain.OutcomeCreated {
		t.Fatalf("= %v, want created", report.Results[0].Outcome)
	}

	f, found, err := flowstorage.NewRepository(testPool).FlowBySlug(ctx, testInstance, "login-mfa")
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	// Order IS meaning for stages; the read must return declaration order.
	want := []domain.StageKind{domain.StageIdentify, domain.StagePassword, domain.StageTOTP}
	for i, k := range want {
		if f.Stages[i].Kind != k {
			t.Fatalf("stage order = %v", f.Stages)
		}
	}

	report, err = eng.Apply(ctx, testPool, testInstance, bp)
	if err != nil {
		t.Fatal(err)
	}
	if report.Changed() {
		t.Fatalf("second apply = %+v, want all-unchanged", report.Results)
	}

	// Reordering is a different flow — updated, not unchanged.
	bp.Entries[0].Attrs["stages"] = []any{
		map[string]any{"kind": "identify"},
		map[string]any{"kind": "totp"},
		map[string]any{"kind": "password"},
	}
	report, err = eng.Apply(ctx, testPool, testInstance, bp)
	if err != nil {
		t.Fatal(err)
	}
	if report.Results[0].Outcome != domain.OutcomeUpdated {
		t.Fatalf("reorder = %v, want updated — swapping password and totp is a different flow", report.Results[0].Outcome)
	}
}

// The proof only a multi-model file can give: entry 1 (a flow) succeeds,
// entry 2 (a seat for a user that does not exist) hits a real constraint —
// and the flow does not survive the rollback either. One file, one fate.
func TestApply_CrossModelRollback(t *testing.T) {
	eng := domain.NewBlueprintEngine(NewApplier(), flowstorage.NewApplier())
	ctx := context.Background()

	_, err := eng.Apply(ctx, testPool, testInstance, &domain.Blueprint{
		Schema: domain.BlueprintSchema,
		Entries: []domain.Entry{
			flowEntry("doomed-flow", map[string]any{"kind": "identify"}, map[string]any{"kind": "password"}),
			seatEntry("missing", "usage", "ws-0001"),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "rolled back") {
		t.Fatalf("err = %v", err)
	}
	_, found, err := flowstorage.NewRepository(testPool).FlowBySlug(ctx, testInstance, "doomed-flow")
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("the flow from entry 1 survived entry 2's rollback — models do not share a fate")
	}
}

func TestExecutions_RoundTrip(t *testing.T) {
	repo := flowstorage.NewRepository(testPool)
	ctx := context.Background()

	exec := &domain.Execution{
		ID: "exec-1", Token: "tok-1", InstanceID: testInstance,
		Plan: &domain.Plan{FlowSlug: "login-password", Designation: domain.DesignationAuthentication,
			Stages: []domain.FlowStage{{Kind: domain.StageIdentify}, {Kind: domain.StagePassword}}},
		ExpiresAt: time.Now().Add(10 * time.Minute).UTC(),
	}
	if err := repo.SaveExecution(ctx, exec); err != nil {
		t.Fatal(err)
	}

	got, found, err := repo.ExecutionByID(ctx, testInstance, "exec-1")
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	if got.Token != "tok-1" || len(got.Plan.Stages) != 2 || got.Position != 0 {
		t.Fatalf("loaded %+v", got)
	}

	got.Position = 1
	got.UserID, got.SessionID = "m1", "sess-9"
	if err := repo.UpdateExecution(ctx, got); err != nil {
		t.Fatal(err)
	}
	got, _, _ = repo.ExecutionByID(ctx, testInstance, "exec-1")
	if got.Position != 1 || got.SessionID != "sess-9" {
		t.Fatalf("update lost: %+v", got)
	}

	if err := repo.DeleteExecution(ctx, testInstance, "exec-1"); err != nil {
		t.Fatal(err)
	}
	if _, found, _ := repo.ExecutionByID(ctx, testInstance, "exec-1"); found {
		t.Fatal("completion consumes the execution; it must not be answerable again")
	}
}
