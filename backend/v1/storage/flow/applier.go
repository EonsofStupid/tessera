package flow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/shippinAI/nomen/backend/v3/storage/database"
)

// FlowModel is what a blueprint entry names to reach this applier.
const FlowModel = "nomen/flow"

// Applier converges `nomen/flow` entries — login configuration as reviewed
// YAML, through the same engine and the same transaction discipline as seats.
type Applier struct{}

func NewApplier() *Applier { return &Applier{} }

var _ domain.BlueprintApplier = (*Applier)(nil)

func (*Applier) Model() string { return FlowModel }

// flowAttrs is the entry's attrs, decoded strictly — same rule, same reason
// as seats: a typo'd `stage:` that silently vanished would be a login flow
// missing a factor, and nobody would know until an audit.
type flowAttrs struct {
	Title       string `json:"title,omitempty"`
	Designation string `json:"designation"`
	Stages      []struct {
		Kind   string         `json:"kind"`
		Config map[string]any `json:"config,omitempty"`
	} `json:"stages"`
}

func decodeEntry(e domain.Entry) (*domain.Flow, error) {
	slug := e.Identifiers["slug"]
	if slug == "" || len(e.Identifiers) != 1 {
		return nil, fmt.Errorf("%s identifies flows by exactly {slug: <slug>}, got %v", FlowModel, e.Identifiers)
	}
	raw, err := json.Marshal(e.Attrs)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var attrs flowAttrs
	if err := dec.Decode(&attrs); err != nil {
		return nil, fmt.Errorf("%s attrs for %s: %w", FlowModel, slug, err)
	}

	designation, err := domain.ParseDesignation(attrs.Designation)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", FlowModel, slug, err)
	}
	f := &domain.Flow{Slug: slug, Title: attrs.Title, Designation: designation}
	for _, st := range attrs.Stages {
		f.Stages = append(f.Stages, domain.FlowStage{Kind: domain.StageKind(st.Kind), Config: st.Config})
	}
	// The domain's own rules — unknown kinds, identify-first, one subject —
	// apply at declaration, so a broken flow is a failed apply rather than a
	// stranded login.
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return f, nil
}

// Apply implements [domain.BlueprintApplier].
func (*Applier) Apply(ctx context.Context, tx database.Transaction, instanceID string, e domain.Entry) (domain.Outcome, string, error) {
	desired, err := decodeEntry(e)
	if err != nil {
		return "", "", err
	}
	current, found, err := flowBySlug(ctx, tx, instanceID, desired.Slug)
	if err != nil {
		return "", "", err
	}
	if found && flowsEqual(current, desired) {
		return domain.OutcomeUnchanged, desired.Slug, nil
	}
	if err := upsertFlow(ctx, tx, instanceID, desired); err != nil {
		return "", "", err
	}
	if found {
		return domain.OutcomeUpdated, desired.Slug, nil
	}
	return domain.OutcomeCreated, desired.Slug, nil
}

// Remove implements [domain.BlueprintApplier].
func (*Applier) Remove(ctx context.Context, tx database.Transaction, instanceID string, e domain.Entry) (domain.Outcome, error) {
	slug := e.Identifiers["slug"]
	if slug == "" {
		return "", fmt.Errorf("%s identifies flows by {slug: <slug>}, got %v", FlowModel, e.Identifiers)
	}
	gone, err := removeFlow(ctx, tx, instanceID, slug)
	if err != nil {
		return "", err
	}
	if gone == 0 {
		return domain.OutcomeUnchanged, nil
	}
	return domain.OutcomeRemoved, nil
}

// flowsEqual compares by declaration. Stage order IS meaning here — unlike a
// seat's workspaces — so the comparison is positional on purpose: swapping
// password and totp is a different flow.
func flowsEqual(a, b *domain.Flow) bool {
	if a.Title != b.Title || a.Designation != b.Designation || len(a.Stages) != len(b.Stages) {
		return false
	}
	for i := range a.Stages {
		if a.Stages[i].Kind != b.Stages[i].Kind {
			return false
		}
		ac, _ := json.Marshal(normalizeConfig(a.Stages[i].Config))
		bc, _ := json.Marshal(normalizeConfig(b.Stages[i].Config))
		if !bytes.Equal(ac, bc) {
			return false
		}
	}
	return true
}

// normalizeConfig makes nil and empty the same thing, which they are.
func normalizeConfig(c map[string]any) map[string]any {
	if c == nil {
		return map[string]any{}
	}
	return c
}
