package domain

import (
	"strings"
	"testing"
)

// A valid blueprint, used as the base every refusal case mutates. Building
// refusals by breaking a known-good file keeps each test about exactly one
// rule.
func validBlueprint() *Blueprint {
	return &Blueprint{
		Schema: BlueprintSchema,
		Name:   "test",
		Entries: []Entry{
			{
				Model:       "tessera/seat",
				ID:          "probe",
				Identifiers: map[string]string{"member": "mem_1"},
				Attrs: map[string]any{
					"occupant":   "agent",
					"basis":      "subscription",
					"workspaces": []any{"ws-0001"},
				},
			},
			{
				Model:       "tessera/grant",
				Identifiers: map[string]string{"seat": "${keyof:probe}", "workspace": "ws-0001"},
			},
		},
	}
}

func TestValidate_AValidBlueprintPasses(t *testing.T) {
	if err := validBlueprint().Validate(); err != nil {
		t.Fatalf("valid blueprint refused: %v", err)
	}
}

// Empty is coherent: a file that declares nothing asks for nothing.
func TestValidate_EmptyEntriesIsValid(t *testing.T) {
	b := &Blueprint{Schema: BlueprintSchema}
	if err := b.Validate(); err != nil {
		t.Fatalf("empty blueprint refused: %v", err)
	}
}

// Every refusal, one table. `want` fragments assert the error is the
// actionable one — naming entries, ids and the fix — not merely non-nil.
func TestValidate_Refusals(t *testing.T) {
	cases := map[string]struct {
		mutate func(*Blueprint)
		want   []string
	}{
		"missing schema": {
			func(b *Blueprint) { b.Schema = "" },
			[]string{`want "tessera.blueprint.v1"`},
		},
		"wrong schema": {
			func(b *Blueprint) { b.Schema = "tessera.blueprint.v2" },
			[]string{`schema is "tessera.blueprint.v2"`},
		},
		"entry without a model": {
			func(b *Blueprint) { b.Entries[0].Model = "  " },
			[]string{"entry 1", "has no model"},
		},
		"entry without identifiers": {
			func(b *Blueprint) { b.Entries[0].Identifiers = nil },
			[]string{"entry 1", "no identifiers", "re-applying would duplicate"},
		},
		"unknown state": {
			func(b *Blueprint) { b.Entries[0].State = "presnt" },
			[]string{"entry 1", `unknown state "presnt"`},
		},
		"absent entry carrying attrs": {
			func(b *Blueprint) { b.Entries[0].State = StateAbsent },
			[]string{"entry 1", "declares absent but carries 3 attrs"},
		},
		"duplicate ids": {
			func(b *Blueprint) { b.Entries[1].ID = "probe" },
			[]string{"entry 2", `reuses id "probe"`, "entry 1"},
		},
		"reference to nothing": {
			func(b *Blueprint) {
				b.Entries[1].Identifiers["seat"] = "${keyof:ghost}"
			},
			[]string{"entry 2", "${keyof:ghost}", "no entry has that id"},
		},
		"empty reference": {
			func(b *Blueprint) {
				b.Entries[1].Identifiers["seat"] = "${keyof:}"
			},
			[]string{"entry 2", "empty reference"},
		},
		"self reference": {
			func(b *Blueprint) {
				b.Entries[0].Attrs["note"] = "${keyof:probe}"
			},
			[]string{"entry 1", "references itself"},
		},
		// The one the plan promised would name both entries and the fix.
		"forward reference": {
			func(b *Blueprint) {
				b.Entries[0].Attrs["grant"] = "${keyof:later}"
				b.Entries[1].ID = "later"
			},
			[]string{"entry 1", "${keyof:later}", "entry 2", "backward only", "move that entry above"},
		},
		"reference nested deep in attrs": {
			func(b *Blueprint) {
				b.Entries[0].Attrs["nested"] = map[string]any{
					"list": []any{map[string]any{"ref": "${keyof:ghost}"}},
				}
			},
			[]string{"entry 1", "${keyof:ghost}", "no entry has that id"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			b := validBlueprint()
			tc.mutate(b)
			err := b.Validate()
			if err == nil {
				t.Fatal("accepted a blueprint that must be refused")
			}
			for _, frag := range tc.want {
				if !strings.Contains(err.Error(), frag) {
					t.Errorf("error lacks %q:\n%v", frag, err)
				}
			}
		})
	}
}

// Three mistakes cost one round trip, not three.
func TestValidate_ReportsEveryFindingAtOnce(t *testing.T) {
	b := validBlueprint()
	b.Schema = ""
	b.Entries[0].Model = ""
	b.Entries[1].Identifiers = nil
	err := b.Validate()
	if err == nil {
		t.Fatal("accepted")
	}
	for _, frag := range []string{"schema", "has no model", "no identifiers"} {
		if !strings.Contains(err.Error(), frag) {
			t.Errorf("first-error-only behaviour: finding %q missing from:\n%v", frag, err)
		}
	}
}

func TestParseState(t *testing.T) {
	for raw, want := range map[string]DesiredState{
		"": StatePresent, "present": StatePresent, " Present ": StatePresent,
		"absent": StateAbsent, "ABSENT": StateAbsent,
	} {
		got, err := ParseState(raw)
		if err != nil || got != want {
			t.Errorf("ParseState(%q) = %q, %v; want %q", raw, got, err, want)
		}
	}
	// No safe default: `present` would apply a typo, `absent` would delete on
	// one.
	for _, raw := range []string{"presnt", "deleted", "created", "yes"} {
		if _, err := ParseState(raw); err == nil {
			t.Errorf("ParseState(%q) guessed instead of refusing", raw)
		}
	}
}
