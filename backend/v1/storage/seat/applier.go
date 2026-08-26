package seat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/shippinAI/nomen/backend/v3/storage/database"
)

// SeatModel is what a blueprint entry names to reach this applier.
const SeatModel = "nomen/seat"

// Applier converges `nomen/seat` entries. It runs inside the engine's
// transaction — the tx-scoped helpers in postgres.go are the same code the
// repository's own methods use, so a blueprint and `nomen seat set` cannot
// disagree about what writing a seat means.
type Applier struct{}

func NewApplier() *Applier { return &Applier{} }

var _ domain.BlueprintApplier = (*Applier)(nil)

func (*Applier) Model() string { return SeatModel }

// seatAttrs is the entry's attrs, decoded strictly.
//
// Strictly, because a blueprint is reviewed text: a typo'd `workspace:`
// (singular) that json decoding silently dropped would apply clean and grant
// nothing — a blueprint that lies. DisallowUnknownFields turns the typo into
// an error naming the field at validate time, which is a review comment.
type seatAttrs struct {
	Account       string   `json:"account"`
	Occupant      string   `json:"occupant,omitempty"`
	Basis         string   `json:"basis,omitempty"`
	Workspaces    []string `json:"workspaces,omitempty"`
	Scopes        []string `json:"scopes,omitempty"`
	PolicyVersion string   `json:"policy_version,omitempty"`
}

// decodeEntry turns a validated entry into the seat it declares.
func decodeEntry(e domain.Entry, instanceID string) (*domain.Seat, error) {
	member := e.Identifiers["member"]
	if member == "" || len(e.Identifiers) != 1 {
		return nil, fmt.Errorf("%s identifies seats by exactly {member: <id>}, got %v", SeatModel, e.Identifiers)
	}

	raw, err := json.Marshal(e.Attrs)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var attrs seatAttrs
	if err := dec.Decode(&attrs); err != nil {
		return nil, fmt.Errorf("%s attrs for member %s: %w", SeatModel, member, err)
	}
	if attrs.Account == "" {
		return nil, fmt.Errorf("%s for member %s: `account` is required — a seat without a tenant is nobody's", SeatModel, member)
	}

	// The axes are checked against the canonical spellings and nothing looser.
	// ParseBasis exists to read *storage* charitably; a blueprint is authored
	// text, and `basis: subscriptoin` quietly becoming `unknown` is exactly
	// the lie strict decoding exists to prevent.
	occupant := attrs.Occupant
	if occupant == "" {
		occupant = string(domain.OccupantAgent)
	}
	if o := domain.ParseOccupant(occupant); string(o) != occupant {
		return nil, fmt.Errorf("%s for member %s: occupant %q is not one of human|agent", SeatModel, member, attrs.Occupant)
	}
	basis := attrs.Basis
	if basis == "" {
		basis = string(domain.BasisUnknown)
	}
	if b := domain.ParseBasis(basis); string(b) != basis {
		return nil, fmt.Errorf("%s for member %s: basis %q is not one of subscription|usage|local|unknown", SeatModel, member, attrs.Basis)
	}

	return &domain.Seat{
		MemberID:      member,
		AccountID:     attrs.Account,
		Occupant:      domain.Occupant(occupant),
		Basis:         domain.Basis(basis),
		Workspaces:    attrs.Workspaces,
		Scopes:        attrs.Scopes,
		PolicyVersion: attrs.PolicyVersion,
	}, nil
}

// Apply implements [domain.BlueprintApplier].
func (*Applier) Apply(ctx context.Context, tx database.Transaction, instanceID string, e domain.Entry) (domain.Outcome, string, error) {
	desired, err := decodeEntry(e, instanceID)
	if err != nil {
		return "", "", err
	}

	current, found, err := seatByMember(ctx, tx, instanceID, desired.MemberID)
	if err != nil {
		return "", "", err
	}
	// `unchanged` is measured, not claimed: skip the write when the stored
	// seat already says everything the entry declares. This is what keeps the
	// second run of a blueprint honest and updated_at meaningful.
	if found && seatsEqual(current, desired) {
		return domain.OutcomeUnchanged, desired.MemberID, nil
	}

	if err := upsertSeat(ctx, tx, instanceID, desired); err != nil {
		return "", "", err
	}
	if found {
		return domain.OutcomeUpdated, desired.MemberID, nil
	}
	return domain.OutcomeCreated, desired.MemberID, nil
}

// Remove implements [domain.BlueprintApplier].
func (*Applier) Remove(ctx context.Context, tx database.Transaction, instanceID string, e domain.Entry) (domain.Outcome, error) {
	member := e.Identifiers["member"]
	if member == "" {
		return "", fmt.Errorf("%s identifies seats by {member: <id>}, got %v", SeatModel, e.Identifiers)
	}
	gone, err := removeSeat(ctx, tx, instanceID, member)
	if err != nil {
		return "", err
	}
	if gone == 0 {
		return domain.OutcomeUnchanged, nil
	}
	return domain.OutcomeRemoved, nil
}

// seatsEqual compares what matters, order-insensitively where order carries no
// meaning: workspaces and scopes are sets in every consumer, and the database
// aggregates them in whatever order it pleases.
func seatsEqual(a, b *domain.Seat) bool {
	return a.AccountID == b.AccountID &&
		domain.ParseOccupant(string(a.Occupant)) == domain.ParseOccupant(string(b.Occupant)) &&
		domain.ParseBasis(string(a.Basis)) == domain.ParseBasis(string(b.Basis)) &&
		a.PolicyVersion == b.PolicyVersion &&
		sameSet(a.Workspaces, b.Workspaces) &&
		sameSet(a.Scopes, b.Scopes)
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	as, bs := slices.Clone(a), slices.Clone(b)
	slices.Sort(as)
	slices.Sort(bs)
	return slices.Equal(as, bs)
}
