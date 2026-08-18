package domain

import (
	"context"
	"errors"
	"fmt"

	"github.com/EonsofStupid/tessera/backend/v3/storage/database"
)

// Outcome is what actually happened to one entry.
//
// `unchanged` is measured, never claimed: an applier reads current state and
// skips the write when equal. That is what makes "apply twice, second run
// reports all-unchanged" a real assertion rather than an UPSERT that happened
// not to error — and it keeps updated_at honest.
type Outcome string

const (
	OutcomeCreated   Outcome = "created"
	OutcomeUpdated   Outcome = "updated"
	OutcomeUnchanged Outcome = "unchanged"
	OutcomeRemoved   Outcome = "removed"
)

// BlueprintApplier makes one model's entries true. Implementations live in
// storage; the engine never learns a table name.
type BlueprintApplier interface {
	// Model names what this applier handles: `tessera/seat`.
	Model() string
	// Apply converges the entry's target to its declared state, inside the
	// engine's transaction. The returned id is what later entries see through
	// `${keyof:…}` — empty is valid for models whose results nothing
	// references.
	Apply(ctx context.Context, tx database.Transaction, instanceID string, e Entry) (Outcome, string, error)
	// Remove converges to absent. Removing what is already gone returns
	// `unchanged`, not an error — a desired state reached twice is reached.
	Remove(ctx context.Context, tx database.Transaction, instanceID string, e Entry) (Outcome, error)
}

// EntryResult is the per-entry line of a report.
type EntryResult struct {
	Entry   string // the same describe() string validation errors use
	Model   string
	Outcome Outcome
}

// Report is what one apply did, entry by entry, in order.
type Report struct {
	Results []EntryResult
}

// Changed reports whether anything happened — the second run of the same
// blueprint must return false, and the probe asserts exactly that.
func (r *Report) Changed() bool {
	for _, res := range r.Results {
		if res.Outcome != OutcomeUnchanged {
			return true
		}
	}
	return false
}

// Counts sums outcomes for a one-line summary.
func (r *Report) Counts() map[Outcome]int {
	c := make(map[Outcome]int, 4)
	for _, res := range r.Results {
		c[res.Outcome]++
	}
	return c
}

// blueprintLockClass is the classid half of the advisory lock every apply
// takes. The other half hashes the instance, so two applies to the *same*
// instance serialise — the panel and an operator interleaving is otherwise a
// race both win — while a fleet applies its tenants in parallel.
const blueprintLockClass = 0x7E55

// BlueprintEngine applies blueprints through registered appliers.
type BlueprintEngine struct {
	appliers map[string]BlueprintApplier
}

// NewBlueprintEngine registers appliers by model. A duplicate model is a
// programming error and panics at construction, where the stack points at the
// registration — not at the first apply, where it points at a customer.
func NewBlueprintEngine(appliers ...BlueprintApplier) *BlueprintEngine {
	m := make(map[string]BlueprintApplier, len(appliers))
	for _, a := range appliers {
		if _, dup := m[a.Model()]; dup {
			panic(fmt.Sprintf("blueprint: two appliers registered for model %q", a.Model()))
		}
		m[a.Model()] = a
	}
	return &BlueprintEngine{appliers: m}
}

// Check is Validate plus the one question validation cannot answer alone:
// does every model have an applier here. Split from Validate so a context
// with no appliers (an editor, the panel) can still validate structure.
func (eng *BlueprintEngine) Check(b *Blueprint) error {
	errs := []error{b.Validate()}
	for i := range b.Entries {
		e := &b.Entries[i]
		if _, ok := eng.appliers[e.Model]; !ok && e.Model != "" {
			errs = append(errs, fmt.Errorf("%s: no applier is registered for model %q", describe(i, e), e.Model))
		}
	}
	return errors.Join(errs...)
}

// Apply makes the whole blueprint true, or none of it.
//
// One transaction for the file. A blueprint that fails on its last entry
// leaves nothing behind — which is what makes it safe to run unattended, and
// is the property 3.3b proves against a real database.
func (eng *BlueprintEngine) Apply(ctx context.Context, db database.Beginner, instanceID string, b *Blueprint) (_ *Report, err error) {
	if instanceID == "" {
		return nil, errors.New("blueprint: an instance to apply to is required")
	}
	// Refuse before Begin: an invalid blueprint should not cost a connection,
	// and "validation failed" arriving wrapped in transaction noise buries the
	// part the author needs to read.
	if err := eng.Check(b); err != nil {
		return nil, err
	}

	tx, err := db.Begin(ctx, &database.TransactionOptions{
		IsolationLevel: database.IsolationLevelReadCommitted,
		AccessMode:     database.AccessModeReadWrite,
	})
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.End(ctx, err) }()

	// Serialise applies per instance, first thing inside the transaction. The
	// lock releases with the transaction, so there is no unlock to forget.
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1, hashtext($2))", blueprintLockClass, instanceID); err != nil {
		return nil, fmt.Errorf("blueprint: could not take the apply lock for %s: %w", instanceID, err)
	}

	report := &Report{Results: make([]EntryResult, 0, len(b.Entries))}
	// What each id produced, for ${keyof:…} in later entries.
	results := map[string]string{}
	// Ids whose entries were declared absent — a later ref to one is a
	// contradiction only the apply can see, and it should say so rather than
	// substitute an empty string into an identifier.
	removed := map[string]bool{}

	for i := range b.Entries {
		e := b.Entries[i] // copy: substitution must not mutate the blueprint

		if err = substituteRefs(&e, i, b, results, removed); err != nil {
			return nil, err
		}

		applier := eng.appliers[e.Model]
		state, _ := ParseState(string(e.State)) // Check already refused unknowns

		var outcome Outcome
		var producedID string
		switch state {
		case StateAbsent:
			outcome, err = applier.Remove(ctx, tx, instanceID, e)
			if e.ID != "" {
				removed[e.ID] = true
			}
		default:
			outcome, producedID, err = applier.Apply(ctx, tx, instanceID, e)
			if e.ID != "" {
				results[e.ID] = producedID
			}
		}
		if err != nil {
			// The rollback in End undoes entries 1..i-1; the error says so,
			// because "did the first half apply?" is the first question an
			// operator asks and the answer should not require a query.
			return nil, fmt.Errorf("%s failed — the transaction rolled back, nothing was applied: %w", describe(i, &b.Entries[i]), err)
		}
		report.Results = append(report.Results, EntryResult{Entry: describe(i, &b.Entries[i]), Model: e.Model, Outcome: outcome})
	}
	return report, nil
}

// substituteRefs resolves every ${keyof:…} in one entry against results so
// far. Validation already guaranteed the targets exist and sit earlier; what
// only apply-time knows is whether a target was removed or produced nothing.
func substituteRefs(e *Entry, i int, b *Blueprint, results map[string]string, removed map[string]bool) error {
	resolve := func(s string) (string, error) {
		var rerr error
		out := refPattern.ReplaceAllStringFunc(s, func(m string) string {
			id := refPattern.FindStringSubmatch(m)[1]
			if removed[id] {
				rerr = fmt.Errorf("%s references ${keyof:%s}, but that entry is declared absent — a removed thing has no id to give", describe(i, e), id)
				return m
			}
			v, ok := results[id]
			if !ok || v == "" {
				rerr = fmt.Errorf("%s references ${keyof:%s}, but that entry produced no id", describe(i, e), id)
				return m
			}
			return v
		})
		return out, rerr
	}

	ids := make(map[string]string, len(e.Identifiers))
	for k, v := range e.Identifiers {
		rv, err := resolve(v)
		if err != nil {
			return err
		}
		ids[k] = rv
	}
	e.Identifiers = ids

	var walk func(v any) (any, error)
	walk = func(v any) (any, error) {
		switch t := v.(type) {
		case string:
			return resolve(t)
		case map[string]any:
			out := make(map[string]any, len(t))
			for k, vv := range t {
				rv, err := walk(vv)
				if err != nil {
					return nil, err
				}
				out[k] = rv
			}
			return out, nil
		case []any:
			out := make([]any, len(t))
			for j, vv := range t {
				rv, err := walk(vv)
				if err != nil {
					return nil, err
				}
				out[j] = rv
			}
			return out, nil
		default:
			return v, nil
		}
	}
	if len(e.Attrs) > 0 {
		attrs, err := walk(map[string]any(e.Attrs))
		if err != nil {
			return err
		}
		e.Attrs = attrs.(map[string]any)
	}
	return nil
}
