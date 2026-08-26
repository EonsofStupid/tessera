package seat

import (
	"strings"
	"testing"

	"github.com/shippinAI/nomen/backend/v1/domain"
)

// The SQL paths are proven against a real Postgres in the atomicity test;
// what lives here is the part that needs no database — the strict decode that
// keeps a blueprint from lying, and the comparison that makes `unchanged` a
// measurement.

func entryWith(attrs map[string]any) domain.Entry {
	return domain.Entry{
		Model:       SeatModel,
		Identifiers: map[string]string{"member": "mem_1"},
		Attrs:       attrs,
	}
}

func TestDecodeEntry_FullSeat(t *testing.T) {
	s, err := decodeEntry(entryWith(map[string]any{
		"account":        "acc_1",
		"occupant":       "human",
		"basis":          "subscription",
		"workspaces":     []any{"ws-0001", "ws-0002"},
		"scopes":         []any{"hosting:active"},
		"policy_version": "pol_1",
	}), "inst-1")
	if err != nil {
		t.Fatal(err)
	}
	if s.MemberID != "mem_1" || s.AccountID != "acc_1" ||
		s.Occupant != domain.OccupantHuman || s.Basis != domain.BasisSubscription ||
		len(s.Workspaces) != 2 || s.PolicyVersion != "pol_1" {
		t.Fatalf("decoded %+v", s)
	}
}

func TestDecodeEntry_DefaultsAreTheSafeOnes(t *testing.T) {
	s, err := decodeEntry(entryWith(map[string]any{"account": "acc_1"}), "inst-1")
	if err != nil {
		t.Fatal(err)
	}
	if s.Occupant != domain.OccupantAgent || s.Basis != domain.BasisUnknown {
		t.Fatalf("defaults = %s/%s, want agent/unknown", s.Occupant, s.Basis)
	}
}

func TestDecodeEntry_Refusals(t *testing.T) {
	cases := map[string]struct {
		entry domain.Entry
		want  string
	}{
		// A typo'd attr silently dropped is a blueprint that applies clean and
		// grants nothing.
		"unknown attr": {
			entryWith(map[string]any{"account": "acc_1", "workspace": []any{"ws-0001"}}),
			`unknown field "workspace"`,
		},
		"missing account": {
			entryWith(map[string]any{"occupant": "agent"}),
			"`account` is required",
		},
		// ParseBasis reads *storage* charitably; a blueprint is authored text,
		// and a misspelling quietly becoming `unknown` is the lie strict
		// decoding exists to prevent.
		"misspelled basis": {
			entryWith(map[string]any{"account": "acc_1", "basis": "subscriptoin"}),
			`basis "subscriptoin" is not one of`,
		},
		"misspelled occupant": {
			entryWith(map[string]any{"account": "acc_1", "occupant": "robot"}),
			`occupant "robot" is not one of`,
		},
		"wrong identifier key": {
			domain.Entry{Model: SeatModel, Identifiers: map[string]string{"user": "u1"}, Attrs: map[string]any{"account": "a"}},
			"exactly {member: <id>}",
		},
		"extra identifier": {
			domain.Entry{Model: SeatModel, Identifiers: map[string]string{"member": "m1", "org": "o1"}, Attrs: map[string]any{"account": "a"}},
			"exactly {member: <id>}",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := decodeEntry(tc.entry, "inst-1")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestSeatsEqual_OrderCarriesNoMeaning(t *testing.T) {
	a := &domain.Seat{AccountID: "acc", Occupant: "agent", Basis: "subscription",
		Workspaces: []string{"ws-0002", "ws-0001"}, Scopes: []string{"b", "a"}}
	b := &domain.Seat{AccountID: "acc", Occupant: "agent", Basis: "subscription",
		Workspaces: []string{"ws-0001", "ws-0002"}, Scopes: []string{"a", "b"}}
	if !seatsEqual(a, b) {
		t.Fatal("the database aggregates in whatever order it pleases; comparison must not care")
	}
	b.Workspaces = []string{"ws-0001"}
	if seatsEqual(a, b) {
		t.Fatal("a different workspace set is a different seat")
	}
}
