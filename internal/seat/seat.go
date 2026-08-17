// Package seat mints the claim set of `shippin.seat-token.v1`.
//
// It is deliberately free of any OIDC or eventstore import. The token contract
// (`docs/01-seat-token-contract.md`) is the product boundary, and a boundary
// that can only be exercised by standing up a provider is one nobody exercises.
// Everything here takes plain facts and returns plain claims, so the rules that
// actually matter — `unknown` is never promoted, `aud` names exactly one
// workspace — are unit tests rather than integration hopes.
//
// The provider behind the contract stays swappable because this package is the
// only thing that knows what a seat token looks like. `internal/api/oidc`
// gathers facts and hands them over; it does not build claims.
package seat

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Schema is the wire identifier every consumer checks before reading anything
// else. A token without it is not a seat token, whatever else it contains.
const Schema = "shippin.seat-token.v1"

// Occupant is who is in the chair — the customer-legible half of "non-human
// identity" (`NAMING-AND-SSOT.md` §1.1).
type Occupant string

const (
	OccupantHuman Occupant = "human"
	OccupantAgent Occupant = "agent"
)

// Basis is how a seat's capacity is paid for. The axis is independent of
// Occupant: an agent seat on a subscription and a human seat on a subscription
// differ in one claim, not in kind.
type Basis string

const (
	BasisSubscription Basis = "subscription"
	BasisUsage        Basis = "usage"
	BasisLocal        Basis = "local"
	// BasisUnknown means nobody measured it. It is a real value, not a null,
	// and it is never promoted to anything else — see ParseBasis.
	BasisUnknown Basis = "unknown"
)

// basisSpellings is every spelling of the basis axis that exists in the estate.
//
// The axis was invented twice: Automaton's adapters write
// `subscription_oauth | api_key | local | unknown`, and the seam draft writes
// `subscription | api | local | none`. Both are read here so neither side has
// to migrate before the other, which is the same accommodation
// `automaton/engine/abstract/seat.mjs` makes in the other direction.
var basisSpellings = map[string]Basis{
	// Automaton's internal spelling.
	"subscription_oauth": BasisSubscription,
	"api_key":            BasisUsage,
	// The seam draft's spelling.
	"subscription": BasisSubscription,
	"api":          BasisUsage,
	"none":         BasisUnknown,
	// Shared by both.
	"local":   BasisLocal,
	"unknown": BasisUnknown,
	// The canonical spelling, which the seam does not use but the contract does.
	"usage": BasisUsage,
}

// ParseBasis maps any known spelling onto the canonical one.
//
// **Anything unrecognised is `unknown`, and `unknown` is never promoted.** This
// is law rather than caution: a basis nobody measured is not a subscription,
// and guessing is how a per-usage bill arrives that nobody chose. The function
// has no default beyond that and takes no "assume" parameter, because the one
// place a caller would reach for it is the place it must not exist.
func ParseBasis(raw string) Basis {
	if b, ok := basisSpellings[strings.TrimSpace(strings.ToLower(raw))]; ok {
		return b
	}
	return BasisUnknown
}

// ParseOccupant maps a stored value onto the axis. Unlike Basis there is no
// safe-by-default direction here — a seat is occupied by someone — so anything
// unrecognised is `agent`, which is the conservative reading: an agent seat is
// the one this estate issues without a person having signed in.
func ParseOccupant(raw string) Occupant {
	if strings.EqualFold(strings.TrimSpace(raw), string(OccupantHuman)) {
		return OccupantHuman
	}
	return OccupantAgent
}

// Authorization is lifted verbatim from `SHARED-SEAM-DRAFT.md` rather than
// invented, so a token and a seam envelope describe entitlement with the same
// words and nothing has to be translated between them.
type Authorization struct {
	Subject       string   `json:"subject"`
	Scopes        []string `json:"scopes"`
	PolicyVersion string   `json:"policy_version,omitempty"`
}

// Provider carries the same axis as Basis under the seam's name for it. It is
// duplication on purpose: the seam envelope already has this field and a
// consumer reading one should not have to know about the other.
type Provider struct {
	AccessClass Basis `json:"access_class"`
}

// Actor is RFC 8693's `act` — delegation, and explicitly not impersonation.
//
// Under impersonation the actor is "indistinguishable" from the subject; under
// delegation it "still has its own identity". `PRODUCT-ARCHITECTURE.md` §7
// requires a persistent visible indicator and a complete audit trail for
// act-as-you mode, and both are only possible under the second — an
// impersonated session has nothing left to indicate.
type Actor struct {
	Subject  string   `json:"sub"`
	Occupant Occupant `json:"occupant,omitempty"`
	// Actor nests to express a chain, outermost being the current actor.
	Actor *Actor `json:"act,omitempty"`
}

// Claims is the Tessera-specific half of a seat token. The registered JWT
// claims (`iss`, `sub`, `aud`, `exp`, `iat`, `nbf`, `jti`) are the OIDC
// layer's to set; this is everything the contract adds on top.
type Claims struct {
	Schema      string   `json:"schema"`
	AccountID   string   `json:"account_id"`
	MemberID    string   `json:"member_id"`
	WorkspaceID string   `json:"workspace_id"`
	Occupant    Occupant `json:"occupant"`
	Basis       Basis    `json:"basis"`

	Authorization Authorization `json:"authorization"`
	Provider      Provider      `json:"provider"`

	Actor *Actor `json:"act,omitempty"`
}

// Facts is what the provider knows at mint time. Every field is something the
// OIDC layer already has; nothing here requires a new query.
type Facts struct {
	// MemberID is the subject — who the token is about, even when an agent is
	// the one acting.
	MemberID string
	// AccountID is the tenant. Today this is the resource-owner organization.
	AccountID string
	// Audience is the token's `aud` exactly as the session recorded it. The
	// workspace is derived from it rather than carried alongside it, so the
	// two cannot disagree.
	Audience []string
	// Occupant and Basis come from the seat's stored facts. An empty Basis is
	// `unknown`, which is a valid answer and the correct one when nothing has
	// been measured.
	Occupant Occupant
	Basis    Basis
	// Scopes is the entitlement, in either the colon or the dotted spelling.
	Scopes []string
	// PolicyVersion names the decision the scopes cite. A scope without one is
	// an entitlement nobody can trace back to a policy.
	PolicyVersion string
	// Actor is set only for a delegated token.
	Actor *Actor
}

// The two ways an audience can fail to name exactly one workspace. They are
// separate errors because the caller must treat them differently: a token whose
// audience names *no* workspace is simply not a seat token — an ordinary OIDC
// client asking for an ordinary token — and minting it unchanged is correct. A
// token naming *two* is a seat token with the tenant boundary missing, and
// there is no safe way to mint that.
var (
	ErrNoWorkspaceAudience        = errors.New("seat: audience names no workspace")
	ErrAmbiguousWorkspaceAudience = errors.New("seat: audience names more than one workspace")
)

// Mint turns facts into the contract's claim set, or refuses.
//
// It refuses rather than degrades. A seat token that cannot say which workspace
// it is for is not a weaker token, it is the multi-tenant boundary missing —
// and the failure mode of minting it anyway is the worst kind: the wrong
// customer's token opens a door and nothing anywhere says so.
func Mint(f Facts) (*Claims, error) {
	workspace, err := WorkspaceFromAudience(f.Audience)
	if err != nil {
		return nil, err
	}
	if f.MemberID == "" {
		return nil, errors.New("seat: a token needs a subject")
	}

	basis := f.Basis
	if basis == "" {
		basis = BasisUnknown
	} else {
		// Round-trip through the parser so an unrecognised value that reached
		// us from storage lands on `unknown` here rather than on the wire.
		basis = ParseBasis(string(basis))
	}
	occupant := f.Occupant
	if occupant == "" {
		occupant = OccupantAgent
	} else {
		occupant = ParseOccupant(string(occupant))
	}

	return &Claims{
		Schema:      Schema,
		AccountID:   f.AccountID,
		MemberID:    f.MemberID,
		WorkspaceID: workspace,
		Occupant:    occupant,
		Basis:       basis,
		Authorization: Authorization{
			Subject:       f.MemberID,
			Scopes:        NormalizeScopes(f.Scopes),
			PolicyVersion: f.PolicyVersion,
		},
		Provider: Provider{AccessClass: basis},
		Actor:    f.Actor,
	}, nil
}

// WorkspaceFromAudience derives `workspace_id` from `aud`.
//
// The contract left open (§Open 3) whether `workspace_id` and `aud` could ever
// disagree. They cannot, because the workspace is read out of the audience
// rather than carried next to it — which settles the question by construction
// instead of by a rule somebody has to remember.
//
// An audience entry is `<service>:<workspace>` (`automaton:ws-0001`), or a bare
// workspace id. Every entry must name the same workspace: a token audible to
// two workspaces is the tenant boundary with a hole in it.
func WorkspaceFromAudience(audience []string) (string, error) {
	seen := ""
	for _, a := range audience {
		ws := workspaceOf(a)
		if ws == "" {
			continue
		}
		if seen != "" && seen != ws {
			return "", fmt.Errorf("%w: %q names both %s and %s",
				ErrAmbiguousWorkspaceAudience, strings.Join(audience, ", "), seen, ws)
		}
		seen = ws
	}
	if seen == "" {
		return "", fmt.Errorf("%w: %q", ErrNoWorkspaceAudience, strings.Join(audience, ", "))
	}
	return seen, nil
}

// workspaceOf reads the workspace half of one audience entry.
//
// Zitadel puts project ids and client ids in `aud` as a matter of course, and
// those are numeric. Requiring the `ws-` prefix is what keeps a project id from
// being mistaken for a workspace and silently becoming the tenant boundary.
func workspaceOf(entry string) string {
	entry = strings.TrimSpace(entry)
	if i := strings.LastIndex(entry, ":"); i >= 0 {
		entry = entry[i+1:]
	}
	if !strings.HasPrefix(entry, "ws-") || entry == "ws-" {
		return ""
	}
	return entry
}

// NormalizeScopes puts entitlement in the canonical spelling and makes the set
// stable.
//
// DevForge's `03-control-plane-contract.md` names these as dotted booleans
// (`hosting.active`); the colon form is canonical because it is what OAuth
// scope syntax and the seam draft already use. Consumers accept both, so this
// is about the issuer being consistent rather than about anyone's compatibility.
//
// Sorted and de-duplicated because two tokens carrying the same entitlement
// should be the same token — an audit trail where the scope order varies by
// map iteration is one nobody can diff.
func NormalizeScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return []string{}
	}
	set := make(map[string]struct{}, len(scopes))
	for _, s := range scopes {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		// Only the first dot becomes a colon: `hosting.active` is
		// `hosting:active`, and a scope with further structure keeps it.
		if !strings.Contains(s, ":") {
			s = strings.Replace(s, ".", ":", 1)
		}
		set[s] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
