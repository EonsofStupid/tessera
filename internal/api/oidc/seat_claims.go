package oidc

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/EonsofStupid/tessera/internal/query"
	"github.com/EonsofStupid/tessera/internal/seat"
)

// Where the seat's facts are stored today.
//
// User metadata, because Zitadel already stores it, already has a management
// API, and is already loaded by the userinfo query — so a seat token costs no
// extra round trip on the token path. It is not meant to be the permanent home:
// Phase 3 (blueprints) is what will write these, and when the source changes,
// only seatFacts below changes with it. Everything downstream takes
// [seat.Facts] and does not know where they came from.
//
// The keys are namespaced so they cannot collide with metadata a customer sets
// for their own purposes.
const (
	metadataKeyOccupant      = "shippin:seat:occupant"
	metadataKeyBasis         = "shippin:seat:basis"
	metadataKeyAccountID     = "shippin:account_id"
	metadataKeyScopes        = "shippin:entitlement:scopes"
	metadataKeyPolicyVersion = "shippin:entitlement:policy_version"
	// Which workspaces this member may occupy, space separated. This is the
	// entitlement behind [domain.SeatAudienceScope]: the scope asks, this
	// decides.
	metadataKeyWorkspaces = "shippin:seat:workspaces"
)

// ErrWorkspaceNotOccupied is the refusal when a caller asks for a workspace the
// member is not entitled to occupy.
//
// It is an error and not a smaller token on purpose. Issuing one without the
// workspace would send a caller away holding something that fails later, at a
// consumer, as a confusing audience mismatch — and the operator would go
// looking at the consumer. Refusing at the mint says which member and which
// workspace, once, in the place that knows both.
var ErrWorkspaceNotOccupied = errors.New("seat: this member does not occupy that workspace")

// rawUserInfoFunc hands back the query result the userinfo closure already
// fetched, so seat facts read from the same row the claims were built from
// rather than from a second query that could disagree with it.
type rawUserInfoFunc func() *query.OIDCUserInfo

// setSeatClaims stamps `shippin.seat-token.v1` onto an access token, or leaves
// it alone.
//
// Two outcomes are both correct and only one is an error:
//
//   - The audience names no workspace. This is an ordinary OIDC client asking
//     for an ordinary token, and it gets one — Tessera is still a working
//     identity provider for things that are not seats. Nothing is stamped.
//   - The audience names exactly one workspace. It is a seat token, and it
//     carries the contract's claims.
//
// An audience naming *two* workspaces is the third case and it is refused,
// because that is the multi-tenant boundary missing rather than absent. The
// contract's rule — "a token minted for ws-0001 presented to ws-0002 is a
// forgery attempt" — only holds if the issuer never mints the ambiguous one.
func setSeatClaims(claims *oidc.AccessTokenClaims, qu *query.OIDCUserInfo, audience []string, subject string) error {
	facts, err := seatFacts(qu, audience, subject)
	if err != nil {
		if errors.Is(err, seat.ErrNoWorkspaceAudience) {
			return nil // not a seat token; not a problem
		}
		// RFC 8693's `invalid_target`: the audience the caller asked for is one
		// this issuer will not mint. Anything else here surfaces as
		// `server_error`, which tells an operator that Tessera broke rather
		// than that they asked for a workspace nobody granted them — and they
		// would go read our logs instead of their own configuration.
		return oidc.ErrInvalidTarget().WithParent(err).WithDescription("%s", err.Error())
	}
	seatClaims, err := seat.Mint(facts)
	if err != nil {
		return err
	}

	if claims.Claims == nil {
		claims.Claims = make(map[string]any)
	}
	claims.Claims["schema"] = seatClaims.Schema
	claims.Claims["account_id"] = seatClaims.AccountID
	claims.Claims["member_id"] = seatClaims.MemberID
	claims.Claims["workspace_id"] = seatClaims.WorkspaceID
	claims.Claims["occupant"] = string(seatClaims.Occupant)
	claims.Claims["basis"] = string(seatClaims.Basis)
	claims.Claims["authorization"] = seatClaims.Authorization
	claims.Claims["provider"] = seatClaims.Provider

	// `act` is left to the OIDC layer, which already builds RFC 8693's chain
	// from the token-exchange actor and marshals it correctly. We add the one
	// claim the contract puts on top of the RFC, and only that — rebuilding the
	// chain here would mean two implementations of it disagreeing later.
	//
	// Everything this estate delegates *through* is an agent seat; a human
	// acting for another human is not a case that exists, and if it ever does
	// it should arrive as a stored fact rather than as a default changed here.
	if claims.Actor != nil {
		annotateActor(claims.Actor)
	}
	return nil
}

func annotateActor(actor *oidc.ActorClaims) {
	if actor == nil {
		return
	}
	if actor.Claims == nil {
		actor.Claims = make(map[string]any)
	}
	actor.Claims["occupant"] = string(seat.OccupantAgent)
	annotateActor(actor.Actor)
}

// seatFacts reads what the provider knows about this seat.
func seatFacts(qu *query.OIDCUserInfo, audience []string, subject string) (seat.Facts, error) {
	md := metadataMap(qu)

	// The gate. A member occupies the workspaces someone recorded, and a
	// request for any other one is refused rather than trimmed — including
	// when nothing was recorded at all, which is an unprovisioned member and
	// not a member with universal access.
	workspace, err := seat.WorkspaceFromAudience(audience)
	if err != nil {
		return seat.Facts{}, err
	}
	if !slices.Contains(strings.Fields(md[metadataKeyWorkspaces]), workspace) {
		return seat.Facts{}, fmt.Errorf("%w: %s is not among %q",
			ErrWorkspaceNotOccupied, workspace, md[metadataKeyWorkspaces])
	}

	// The account is the resource-owner organization unless something says
	// otherwise. `account_id is the tenant today` (contract §Open 4) — the
	// metadata key exists so that stays true when organizations grow a level.
	accountID := md[metadataKeyAccountID]
	if accountID == "" && qu != nil && qu.Org != nil {
		accountID = qu.Org.ID
	}

	return seat.Facts{
		MemberID:      subject,
		AccountID:     accountID,
		Audience:      audience,
		Occupant:      seat.ParseOccupant(md[metadataKeyOccupant]),
		Basis:         seat.ParseBasis(md[metadataKeyBasis]),
		Scopes:        strings.Fields(md[metadataKeyScopes]),
		PolicyVersion: md[metadataKeyPolicyVersion],
	}, nil
}

// metadataMap flattens the user's metadata to plain strings.
//
// Note this is the *stored* value, not the base64 form `setUserInfoMetadata`
// puts in the `urn:zitadel:iam:user:metadata` claim. Seat claims are read from
// storage and re-encoded by this package; they never travel through the
// metadata claim, so a client that has not asked for the metadata scope still
// gets a correct seat token.
func metadataMap(qu *query.OIDCUserInfo) map[string]string {
	if qu == nil || len(qu.Metadata) == 0 {
		return nil
	}
	out := make(map[string]string, len(qu.Metadata))
	for _, md := range qu.Metadata {
		out[md.Key] = string(md.Value)
	}
	return out
}
