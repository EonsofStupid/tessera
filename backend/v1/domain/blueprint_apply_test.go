package domain

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/EonsofStupid/tessera/backend/v3/storage/database"
)

// ---- fakes: a transaction that records, an applier that is scripted --------

type fakeTx struct {
	execs      []string
	execArgs   [][]any
	committed  bool
	rolledBack bool
}

func (t *fakeTx) Exec(_ context.Context, stmt string, args ...any) (int64, error) {
	t.execs = append(t.execs, stmt)
	t.execArgs = append(t.execArgs, args)
	return 0, nil
}
func (t *fakeTx) Query(context.Context, string, ...any) (database.Rows, error) {
	panic("engine must not query directly")
}
func (t *fakeTx) QueryRow(context.Context, string, ...any) database.Row {
	panic("engine must not query directly")
}
func (t *fakeTx) Commit(context.Context) error   { t.committed = true; return nil }
func (t *fakeTx) Rollback(context.Context) error { t.rolledBack = true; return nil }
func (t *fakeTx) End(ctx context.Context, err error) error {
	// The real contract: commit on nil, rollback otherwise.
	if err != nil {
		_ = t.Rollback(ctx)
		return err
	}
	return t.Commit(ctx)
}
func (t *fakeTx) Begin(context.Context) (database.Transaction, error) {
	panic("engine must not nest transactions")
}

type fakeBeginner struct {
	tx     *fakeTx
	begins int
}

func (b *fakeBeginner) Begin(context.Context, *database.TransactionOptions) (database.Transaction, error) {
	b.begins++
	return b.tx, nil
}

// step is one scripted applier response.
type step struct {
	outcome Outcome
	id      string
	err     error
}

type fakeApplier struct {
	model   string
	script  []step
	applied []Entry // what Apply/Remove actually received, post-substitution
	removed []Entry
}

func (a *fakeApplier) Model() string { return a.model }
func (a *fakeApplier) next() step {
	if len(a.script) == 0 {
		return step{outcome: OutcomeCreated, id: "id-" + a.model}
	}
	s := a.script[0]
	a.script = a.script[1:]
	return s
}
func (a *fakeApplier) Apply(_ context.Context, _ database.Transaction, _ string, e Entry) (Outcome, string, error) {
	a.applied = append(a.applied, e)
	s := a.next()
	return s.outcome, s.id, s.err
}
func (a *fakeApplier) Remove(_ context.Context, _ database.Transaction, _ string, e Entry) (Outcome, error) {
	a.removed = append(a.removed, e)
	s := a.next()
	return s.outcome, s.err
}

func engineWith(appliers ...BlueprintApplier) (*BlueprintEngine, *fakeBeginner) {
	return NewBlueprintEngine(appliers...), &fakeBeginner{tx: &fakeTx{}}
}

// ---- the semantics ---------------------------------------------------------

func TestApply_CommitsOnceAndLocksFirst(t *testing.T) {
	seats := &fakeApplier{model: "tessera/seat"}
	eng, db := engineWith(seats)

	report, err := eng.Apply(context.Background(), db, "inst-1", &Blueprint{
		Schema: BlueprintSchema,
		Entries: []Entry{
			{Model: "tessera/seat", Identifiers: map[string]string{"member": "m1"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if db.begins != 1 || !db.tx.committed || db.tx.rolledBack {
		t.Fatalf("begins=%d committed=%v rolledBack=%v — want exactly one committed transaction", db.begins, db.tx.committed, db.tx.rolledBack)
	}
	// The advisory lock is the first statement inside the transaction and is
	// keyed per instance, so tenants apply in parallel and one tenant's two
	// applies cannot interleave.
	if len(db.tx.execs) == 0 || !strings.Contains(db.tx.execs[0], "pg_advisory_xact_lock") {
		t.Fatalf("first statement = %q, want the advisory lock", db.tx.execs)
	}
	if args := db.tx.execArgs[0]; len(args) != 2 || args[1] != "inst-1" {
		t.Fatalf("lock args = %v, want (class, instance)", args)
	}
	if len(report.Results) != 1 || report.Results[0].Outcome != OutcomeCreated {
		t.Fatalf("report = %+v", report.Results)
	}
}

func TestApply_FailureRollsBackAndNamesTheEntry(t *testing.T) {
	seats := &fakeApplier{model: "tessera/seat", script: []step{
		{outcome: OutcomeCreated, id: "s1"},
		{err: errors.New("column does not exist")},
	}}
	eng, db := engineWith(seats)

	_, err := eng.Apply(context.Background(), db, "inst-1", &Blueprint{
		Schema: BlueprintSchema,
		Entries: []Entry{
			{Model: "tessera/seat", ID: "a", Identifiers: map[string]string{"member": "m1"}},
			{Model: "tessera/seat", ID: "b", Identifiers: map[string]string{"member": "m2"}},
		},
	})
	if err == nil {
		t.Fatal("want the applier's failure")
	}
	// The operator's first question is "did the first half apply?" — the error
	// answers it without a query.
	for _, frag := range []string{`entry 2`, `"b"`, "rolled back", "nothing was applied", "column does not exist"} {
		if !strings.Contains(err.Error(), frag) {
			t.Errorf("error lacks %q:\n%v", frag, err)
		}
	}
	if db.tx.committed || !db.tx.rolledBack {
		t.Fatalf("committed=%v rolledBack=%v — want rollback only", db.tx.committed, db.tx.rolledBack)
	}
}

func TestApply_SubstitutesRefsFromEarlierResults(t *testing.T) {
	seats := &fakeApplier{model: "tessera/seat", script: []step{{outcome: OutcomeCreated, id: "seat-777"}}}
	grants := &fakeApplier{model: "tessera/grant"}
	eng, db := engineWith(seats, grants)

	_, err := eng.Apply(context.Background(), db, "inst-1", &Blueprint{
		Schema: BlueprintSchema,
		Entries: []Entry{
			{Model: "tessera/seat", ID: "probe", Identifiers: map[string]string{"member": "m1"}},
			{Model: "tessera/grant",
				Identifiers: map[string]string{"seat": "${keyof:probe}"},
				Attrs:       map[string]any{"nested": []any{map[string]any{"ref": "prefix-${keyof:probe}"}}}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := grants.applied[0]
	if got.Identifiers["seat"] != "seat-777" {
		t.Errorf("identifier = %q, want the produced id", got.Identifiers["seat"])
	}
	nested := got.Attrs["nested"].([]any)[0].(map[string]any)["ref"]
	if nested != "prefix-seat-777" {
		t.Errorf("nested attr = %q — substitution must reach every depth and keep surrounding text", nested)
	}
}

func TestApply_RefToEntryThatProducedNothingIsAnError(t *testing.T) {
	seats := &fakeApplier{model: "tessera/seat", script: []step{{outcome: OutcomeCreated, id: ""}}}
	grants := &fakeApplier{model: "tessera/grant"}
	eng, db := engineWith(seats, grants)

	_, err := eng.Apply(context.Background(), db, "inst-1", &Blueprint{
		Schema: BlueprintSchema,
		Entries: []Entry{
			{Model: "tessera/seat", ID: "probe", Identifiers: map[string]string{"member": "m1"}},
			{Model: "tessera/grant", Identifiers: map[string]string{"seat": "${keyof:probe}"}},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "produced no id") {
		t.Fatalf("err = %v, want 'produced no id'", err)
	}
	if !db.tx.rolledBack {
		t.Error("the failed apply must roll back")
	}
}

func TestApply_RefToARemovedEntryIsAContradiction(t *testing.T) {
	seats := &fakeApplier{model: "tessera/seat", script: []step{{outcome: OutcomeRemoved}}}
	grants := &fakeApplier{model: "tessera/grant"}
	eng, db := engineWith(seats, grants)

	_, err := eng.Apply(context.Background(), db, "inst-1", &Blueprint{
		Schema: BlueprintSchema,
		Entries: []Entry{
			{Model: "tessera/seat", ID: "probe", State: StateAbsent, Identifiers: map[string]string{"member": "m1"}},
			{Model: "tessera/grant", Identifiers: map[string]string{"seat": "${keyof:probe}"}},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "declared absent") {
		t.Fatalf("err = %v, want the contradiction named", err)
	}
	if !db.tx.rolledBack {
		t.Error("must roll back")
	}
}

func TestApply_AbsentGoesThroughRemove(t *testing.T) {
	seats := &fakeApplier{model: "tessera/seat", script: []step{{outcome: OutcomeUnchanged}}}
	eng, db := engineWith(seats)

	report, err := eng.Apply(context.Background(), db, "inst-1", &Blueprint{
		Schema: BlueprintSchema,
		Entries: []Entry{
			{Model: "tessera/seat", State: StateAbsent, Identifiers: map[string]string{"member": "gone"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(seats.removed) != 1 || len(seats.applied) != 0 {
		t.Fatalf("removed=%d applied=%d — absent must route to Remove", len(seats.removed), len(seats.applied))
	}
	// Removing the already-gone is unchanged: reached twice is reached.
	if report.Changed() {
		t.Error("an all-unchanged report must say nothing changed")
	}
	if !db.tx.committed {
		t.Error("an unchanged apply still commits — it held the lock and proved convergence")
	}
}

// The idempotence contract at engine level: scripted all-unchanged is exactly
// what the second run of the same blueprint looks like, and the probe asserts
// Changed() == false against the real database in 3.4a.
func TestApply_SecondRunReportsAllUnchanged(t *testing.T) {
	seats := &fakeApplier{model: "tessera/seat", script: []step{
		{outcome: OutcomeUnchanged, id: "s1"},
		{outcome: OutcomeUnchanged, id: "s2"},
	}}
	eng, db := engineWith(seats)

	report, err := eng.Apply(context.Background(), db, "inst-1", &Blueprint{
		Schema: BlueprintSchema,
		Entries: []Entry{
			{Model: "tessera/seat", Identifiers: map[string]string{"member": "m1"}},
			{Model: "tessera/seat", Identifiers: map[string]string{"member": "m2"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Changed() {
		t.Fatalf("Changed() = true on an all-unchanged run: %+v", report.Results)
	}
	if c := report.Counts(); c[OutcomeUnchanged] != 2 {
		t.Fatalf("counts = %v", c)
	}
	_ = db
}

func TestApply_RefusesBeforeBegin(t *testing.T) {
	eng, db := engineWith(&fakeApplier{model: "tessera/seat"})

	// Unknown model: Check fails, no transaction is opened, no lock is taken.
	_, err := eng.Apply(context.Background(), db, "inst-1", &Blueprint{
		Schema: BlueprintSchema,
		Entries: []Entry{
			{Model: "tessera/ghost", Identifiers: map[string]string{"x": "y"}},
		},
	})
	if err == nil || !strings.Contains(err.Error(), `no applier is registered for model "tessera/ghost"`) {
		t.Fatalf("err = %v", err)
	}
	if db.begins != 0 {
		t.Error("an invalid blueprint must not cost a transaction")
	}

	// No instance: same rule.
	if _, err := eng.Apply(context.Background(), db, "", &Blueprint{Schema: BlueprintSchema}); err == nil {
		t.Fatal("applying to no instance must be refused")
	}
	if db.begins != 0 {
		t.Error("still no transaction")
	}
}

func TestNewBlueprintEngine_DuplicateModelPanicsAtConstruction(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("two appliers for one model must panic where the stack names the registration")
		}
	}()
	NewBlueprintEngine(&fakeApplier{model: "tessera/seat"}, &fakeApplier{model: "tessera/seat"})
}
