package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// BlueprintSchema is the wire identifier a blueprint file must carry. A file
// without it is not a blueprint, whatever else it contains — the same rule the
// seat token applies to its own schema claim.
const BlueprintSchema = "nomen.blueprint.v1"

// DesiredState is what an entry declares should be true of its target.
//
// Two states, on purpose. `absent` of something already gone is success — a
// desired state reached twice is reached. Nomen's must_created guard is
// deferred until something needs it; a state nobody uses is a state nobody
// tests.
type DesiredState string

const (
	StatePresent DesiredState = "present"
	StateAbsent  DesiredState = "absent"
)

// ParseState maps a stored value onto the axis — and unlike ParseBasis it has
// NO safe default. A basis nobody measured is safely `unknown`; a state nobody
// spelled correctly is not safely anything, because guessing `present` would
// apply a typo and guessing `absent` would delete on one. Empty means
// `present` (the unmarked case every blueprint is full of); anything else
// unrecognised is an error.
func ParseState(raw string) (DesiredState, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "", string(StatePresent):
		return StatePresent, nil
	case string(StateAbsent):
		return StateAbsent, nil
	default:
		return "", fmt.Errorf("unknown state %q (want %q or %q)", raw, StatePresent, StateAbsent)
	}
}

// Entry declares one thing that should be true.
type Entry struct {
	// Model names the applier: `nomen/seat`. Whether a registered applier
	// exists for it is the registry's question, not this struct's — validation
	// here is structural, so a blueprint can be validated in a context that has
	// no appliers at all (an editor, a CI check, the panel).
	Model string `json:"model"`
	// ID is a local handle for `${keyof:…}` references from later entries.
	// Optional unless something references it.
	ID string `json:"id,omitempty"`
	// Identifiers are the upsert key: the applier finds the existing object by
	// these, or creates it. They are what make re-applying a no-op instead of
	// a duplicate.
	Identifiers map[string]string `json:"identifiers"`
	// Attrs is everything the entry sets. Appliers decode it strictly — an
	// attr the model does not have is an error, never silently dropped.
	Attrs map[string]any `json:"attrs,omitempty"`
	// State defaults to present.
	State DesiredState `json:"state,omitempty"`
}

// Blueprint is one file's worth of declared truth, applied in one transaction.
type Blueprint struct {
	Schema  string  `json:"schema"`
	Name    string  `json:"name,omitempty"`
	Entries []Entry `json:"entries"`
}

// describe names an entry the way a review comment would: by position, model
// and handle. Every validation error goes through this, because "entry 3
// (nomen/seat \"probe\")" is findable in a file and "invalid entry" is not.
func describe(i int, e *Entry) string {
	if e.ID != "" {
		return fmt.Sprintf("entry %d (%s %q)", i+1, e.Model, e.ID)
	}
	return fmt.Sprintf("entry %d (%s)", i+1, e.Model)
}

// refPattern is the reference syntax: ${keyof:<id>}. A string convention
// rather than a YAML tag, so the same blueprint is writable as JSON — the
// panel generates these, and a format only writable by a YAML library is a
// format the panel fights.
var refPattern = regexp.MustCompile(`\$\{keyof:([^}]*)\}`)

// Validate checks everything that can be known without a registry or a
// database. It returns all findings at once rather than the first — a file
// with three mistakes should cost one round trip to fix, not three.
func (b *Blueprint) Validate() error {
	var errs []error

	if b.Schema != BlueprintSchema {
		errs = append(errs, fmt.Errorf("schema is %q, want %q — refusing to guess at a format nobody declared", b.Schema, BlueprintSchema))
	}

	// Where each local id is defined, by entry index — both for duplicate
	// detection and for telling a forward reference exactly where its target
	// sits.
	definedAt := map[string]int{}
	for i := range b.Entries {
		e := &b.Entries[i]
		if e.ID == "" {
			continue
		}
		if prev, dup := definedAt[e.ID]; dup {
			errs = append(errs, fmt.Errorf("%s reuses id %q already taken by %s — a reference to it would be ambiguous",
				describe(i, e), e.ID, describe(prev, &b.Entries[prev])))
			continue
		}
		definedAt[e.ID] = i
	}

	for i := range b.Entries {
		e := &b.Entries[i]

		if strings.TrimSpace(e.Model) == "" {
			errs = append(errs, fmt.Errorf("%s has no model — nothing can apply it", describe(i, e)))
		}
		if len(e.Identifiers) == 0 {
			errs = append(errs, fmt.Errorf("%s has no identifiers — without an upsert key, re-applying would duplicate instead of converge", describe(i, e)))
		}

		state, err := ParseState(string(e.State))
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", describe(i, e), err))
		}
		if state == StateAbsent && len(e.Attrs) > 0 {
			// Attrs on an absent entry would be silently ignored, and a
			// blueprint that carries settings nothing reads is a blueprint
			// that lies about what it does.
			errs = append(errs, fmt.Errorf("%s declares absent but carries %d attrs — remove them, or remove `state: absent`", describe(i, e), len(e.Attrs)))
		}

		// References reach backward only. Applied in file order, an entry can
		// only use results that already exist; a forward reference is an
		// authoring mistake best caught as a review comment, not a runtime
		// surprise — so the error names both entries and says what to do.
		for _, id := range collectRefs(e) {
			target, defined := definedAt[id]
			switch {
			case id == "":
				errs = append(errs, fmt.Errorf("%s contains an empty reference ${keyof:} — name the entry it means", describe(i, e)))
			case !defined:
				errs = append(errs, fmt.Errorf("%s references ${keyof:%s} but no entry has that id", describe(i, e), id))
			case target == i:
				errs = append(errs, fmt.Errorf("%s references itself — an entry cannot use its own result", describe(i, e)))
			case target > i:
				errs = append(errs, fmt.Errorf("%s references ${keyof:%s}, defined later by %s — references reach backward only; move that entry above this one",
					describe(i, e), id, describe(target, &b.Entries[target])))
			}
		}
	}

	return errors.Join(errs...)
}

// collectRefs finds every ${keyof:…} an entry uses, in identifiers and
// anywhere in attrs — nested maps and lists included, because a reference
// three levels deep is still a dependency.
func collectRefs(e *Entry) []string {
	var refs []string
	for _, v := range e.Identifiers {
		refs = append(refs, refsInString(v)...)
	}
	var walk func(v any)
	walk = func(v any) {
		switch t := v.(type) {
		case string:
			refs = append(refs, refsInString(t)...)
		case map[string]any:
			for _, vv := range t {
				walk(vv)
			}
		case []any:
			for _, vv := range t {
				walk(vv)
			}
		}
	}
	for _, v := range e.Attrs {
		walk(v)
	}
	return refs
}

func refsInString(s string) []string {
	var out []string
	for _, m := range refPattern.FindAllStringSubmatch(s, -1) {
		out = append(out, m[1])
	}
	return out
}
