package oidc

import (
	"context"
	"errors"

	"github.com/shippinAI/nomen/oidc/v3/pkg/oidc"

	// Nomen's own domain layer. Aliased because the package is `domain` and
	// every other file in this package means `internal/domain` by that name —
	// two different domains under one spelling is how the wrong one gets
	// imported at three in the morning.
	seat "github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/shippinAI/nomen/internal/api/authz"
)

// setSeatClaims stamps `shippin.seat-token.v1` onto an access token, or leaves
// it alone.
//
// This is an adapter and holds no rules. It reads a seat through the port and
// asks the seat for a token; which workspaces a member may occupy, what
// `unknown` may become and how an audience maps to a workspace are all decided
// in `backend/v1/domain`, because a rule enforced here would be a rule enforced
// on one mint path out of four.
//
// Two outcomes are both correct and only one is an error:
//
//   - The audience names no workspace. This is an ordinary OIDC client asking
//     for an ordinary token, and it gets one — Nomen is still a working
//     identity provider for things that are not seats. Nothing is stamped.
//   - The audience names exactly one workspace the seat occupies. It carries
//     the contract's claims.
//
// Everything else is refused.
func (s *Server) setSeatClaims(ctx context.Context, claims *oidc.AccessTokenClaims, audience []string, subject string) error {
	// The instance is the tenant root and every seat lookup is scoped by it.
	// Reading it from the request context rather than from the token means a
	// seat in one instance can never answer for a member id that collides with
	// one in another.
	instanceID := authz.GetInstance(ctx).InstanceID()

	seated, err := s.seats.SeatByMember(ctx, instanceID, subject)
	if err != nil {
		return err
	}

	seatClaims, err := seated.Token(audience, nil)
	if err != nil {
		if errors.Is(err, seat.ErrNoWorkspaceAudience) {
			return nil // not a seat token; not a problem
		}
		// RFC 8693's `invalid_target`: the audience the caller asked for is one
		// this issuer will not mint. Anything else surfaces as `server_error`,
		// which tells an operator that Nomen broke rather than that they
		// asked for a workspace nobody granted them — and they would go read
		// our logs instead of their own configuration.
		return oidc.ErrInvalidTarget().WithParent(err).WithDescription("%s", err.Error())
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
	annotateActor(claims.Actor)
	return nil
}

// annotateActor names every actor in a delegation chain an agent.
//
// Everything this estate delegates *through* is an agent seat; a human acting
// for another human is not a case that exists, and if it ever does it should
// arrive as a stored fact rather than as a default changed here.
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
