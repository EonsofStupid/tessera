package seat

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shippinAI/nomen/backend/v1/domain"
	nomenmigration "github.com/shippinAI/nomen/backend/v1/storage/migration"
	"github.com/shippinAI/nomen/backend/v3/storage/database"
	v3postgres "github.com/shippinAI/nomen/backend/v3/storage/database/dialect/postgres"
)

// The proof 3.3b exists for: a blueprint that fails on its last entry leaves
// the database byte-identical, and the second run of a converged blueprint
// reports all-unchanged. Both against a real PostgreSQL 16 — the same engine
// production runs — because a rollback proven against a fake proves the fake.
//
// If the embedded binary is missing and cannot download, the skip names the
// fix: bash dev/preflight.sh --fetch-embedded

var (
	testPool database.Pool
	rawPool  *pgxpool.Pool
)

const testInstance = "inst-atomicity"

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	ctx := context.Background()

	// Started directly rather than through the v3 harness, because that
	// helper hides its connection URL and our tern migrator needs pgx —
	// same library, same PG 16, one extra variable.
	tmp, err := os.MkdirTemp("", "nomen-atomicity-*")
	if err != nil {
		log.Print(err)
		return 1
	}
	defer os.RemoveAll(tmp)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Print(err)
		return 1
	}
	port := uint32(l.Addr().(*net.TCPAddr).Port)
	_ = l.Close()

	cfg := embeddedpostgres.DefaultConfig().Version(embeddedpostgres.V16).
		Port(port).RuntimePath(tmp).StartTimeout(90 * time.Second)
	pg := embeddedpostgres.NewDatabase(cfg)
	if err := pg.Start(); err != nil {
		// Not a failure of the code under test: the environment lacks the
		// binary. The preflight populates it once; until then this proof is
		// unavailable here, and saying so beats failing cryptically.
		log.Printf("SKIP: embedded postgres unavailable (%v) — run: bash dev/preflight.sh --fetch-embedded", err)
		return 0
	}
	defer func() { _ = pg.Stop() }()

	url := cfg.GetConnectionURL() + "?sslmode=disable"

	// nomen schema first — nomen_product.seats carries a foreign key into
	// nomen.users, exactly like production ordering in start.go.
	connector, err := v3postgres.DecodeConfig(url)
	if err != nil {
		log.Print(err)
		return 1
	}
	v3pool, err := connector.Connect(ctx)
	if err != nil {
		log.Print(err)
		return 1
	}
	testPool = v3pool.(database.Pool)
	if err := v3pool.(database.PoolTest).MigrateTest(ctx); err != nil {
		log.Printf("v3 migration: %v", err)
		return 1
	}

	rawPool, err = pgxpool.New(ctx, url)
	if err != nil {
		log.Print(err)
		return 1
	}
	defer rawPool.Close()
	if err := nomenmigration.Migrate(ctx, rawPool); err != nil {
		log.Printf("nomen migration: %v", err)
		return 1
	}

	// The FK targets: one instance, one org, two members. Deliberately NOT a
	// third member — "missing" stays missing so the broken-entry test can hit
	// a real constraint.
	for _, stmt := range []string{
		`INSERT INTO nomen.instances (id, name) VALUES ($1, 'atomicity test')`,
		`INSERT INTO nomen.organizations (id, name, instance_id, state) VALUES ('org-1', 'org', $1, 'active')`,
		`INSERT INTO nomen.users (instance_id, organization_id, id, username, type, name) VALUES ($1, 'org-1', 'm1', 'm1', 'machine', 'm1')`,
		`INSERT INTO nomen.users (instance_id, organization_id, id, username, type, name) VALUES ($1, 'org-1', 'm2', 'm2', 'machine', 'm2')`,
	} {
		if _, err := rawPool.Exec(ctx, stmt, testInstance); err != nil {
			log.Printf("seed: %v", err)
			return 1
		}
	}

	return m.Run()
}

func seatEntry(member, basis string, workspaces ...any) domain.Entry {
	return domain.Entry{
		Model:       SeatModel,
		Identifiers: map[string]string{"member": member},
		Attrs: map[string]any{
			"account":    "org-1",
			"occupant":   "agent",
			"basis":      basis,
			"workspaces": workspaces,
			"scopes":     []any{"hosting:active"},
		},
	}
}

// snapshot renders every nomen row — timestamps included, so a write that
// was rolled back cannot even leave a touched updated_at behind.
func snapshot(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	var b strings.Builder
	for _, q := range []string{
		`SELECT member_id, account_id, occupant::text, basis::text, scopes::text, policy_version,
		        created_at::text, updated_at::text
		 FROM nomen_product.seats ORDER BY instance_id, member_id`,
		`SELECT member_id, workspace_id, created_at::text
		 FROM nomen_product.seat_workspaces ORDER BY instance_id, member_id, workspace_id`,
	} {
		rows, err := rawPool.Query(ctx, q)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			vals, err := rows.Values()
			if err != nil {
				t.Fatal(err)
			}
			fmt.Fprintf(&b, "%v\n", vals)
		}
		rows.Close()
	}
	return b.String()
}

func TestApply_ConvergesAndTheSecondRunChangesNothing(t *testing.T) {
	eng := domain.NewBlueprintEngine(NewApplier())
	bp := &domain.Blueprint{
		Schema: domain.BlueprintSchema,
		Entries: []domain.Entry{
			seatEntry("m1", "subscription", "ws-0001", "ws-0002"),
		},
	}

	report, err := eng.Apply(context.Background(), testPool, testInstance, bp)
	if err != nil {
		t.Fatal(err)
	}
	if report.Results[0].Outcome != domain.OutcomeCreated {
		t.Fatalf("first apply = %v, want created", report.Results[0].Outcome)
	}

	// Through the repository, so what the blueprint wrote is what the token
	// path will read.
	seat, err := NewRepository(testPool).SeatByMember(context.Background(), testInstance, "m1")
	if err != nil {
		t.Fatal(err)
	}
	if seat.Basis != domain.BasisSubscription || len(seat.Workspaces) != 2 {
		t.Fatalf("read back %+v", seat)
	}

	before := snapshot(t)
	report, err = eng.Apply(context.Background(), testPool, testInstance, bp)
	if err != nil {
		t.Fatal(err)
	}
	if report.Changed() {
		t.Fatalf("second run = %+v, want all-unchanged", report.Results)
	}
	if after := snapshot(t); after != before {
		t.Errorf("an all-unchanged run altered the database:\nbefore:\n%s\nafter:\n%s", before, after)
	}

	// Convergence is not append-only: shrinking the declaration shrinks the
	// entitlement, or revoking access becomes something only a human can do.
	bp.Entries[0].Attrs["workspaces"] = []any{"ws-0001"}
	report, err = eng.Apply(context.Background(), testPool, testInstance, bp)
	if err != nil {
		t.Fatal(err)
	}
	if report.Results[0].Outcome != domain.OutcomeUpdated {
		t.Fatalf("shrunk apply = %v, want updated", report.Results[0].Outcome)
	}
	seat, _ = NewRepository(testPool).SeatByMember(context.Background(), testInstance, "m1")
	if len(seat.Workspaces) != 1 || seat.Workspaces[0] != "ws-0001" {
		t.Fatalf("workspaces = %v, want the declared set exactly", seat.Workspaces)
	}
}

func TestApply_BrokenLastEntryLeavesTheDatabaseByteIdentical(t *testing.T) {
	eng := domain.NewBlueprintEngine(NewApplier())
	before := snapshot(t)

	_, err := eng.Apply(context.Background(), testPool, testInstance, &domain.Blueprint{
		Schema: domain.BlueprintSchema,
		Entries: []domain.Entry{
			// Entry 1 is valid and WOULD create m2's seat…
			seatEntry("m2", "usage", "ws-0009"),
			// …entry 2 hits a real constraint: no such user, so the seat's
			// foreign key into nomen.users refuses it.
			seatEntry("missing", "usage", "ws-0009"),
		},
	})
	if err == nil {
		t.Fatal("an entry violating a real constraint must fail the apply")
	}
	for _, frag := range []string{"entry 2", "rolled back", "nothing was applied"} {
		if !strings.Contains(err.Error(), frag) {
			t.Errorf("error lacks %q:\n%v", frag, err)
		}
	}

	if after := snapshot(t); after != before {
		t.Fatalf("the failed apply left a trace — entry 1's seat survived the rollback:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	// And m2 specifically — the entry that succeeded before the failure — must
	// not exist.
	seat, err := NewRepository(testPool).SeatByMember(context.Background(), testInstance, "m2")
	if err != nil {
		t.Fatal(err)
	}
	if len(seat.Workspaces) != 0 {
		t.Fatalf("m2 occupies %v after a rolled-back apply", seat.Workspaces)
	}
}

func TestApply_AbsentRemovesAndRemovingTheGoneIsUnchanged(t *testing.T) {
	eng := domain.NewBlueprintEngine(NewApplier())
	ctx := context.Background()

	if _, err := eng.Apply(ctx, testPool, testInstance, &domain.Blueprint{
		Schema:  domain.BlueprintSchema,
		Entries: []domain.Entry{seatEntry("m2", "local", "ws-0004")},
	}); err != nil {
		t.Fatal(err)
	}

	gone := &domain.Blueprint{
		Schema: domain.BlueprintSchema,
		Entries: []domain.Entry{{
			Model:       SeatModel,
			State:       domain.StateAbsent,
			Identifiers: map[string]string{"member": "m2"},
		}},
	}
	report, err := eng.Apply(ctx, testPool, testInstance, gone)
	if err != nil {
		t.Fatal(err)
	}
	if report.Results[0].Outcome != domain.OutcomeRemoved {
		t.Fatalf("= %v, want removed", report.Results[0].Outcome)
	}
	// Reached twice is reached.
	report, err = eng.Apply(ctx, testPool, testInstance, gone)
	if err != nil {
		t.Fatal(err)
	}
	if report.Results[0].Outcome != domain.OutcomeUnchanged {
		t.Fatalf("= %v, want unchanged", report.Results[0].Outcome)
	}
}
