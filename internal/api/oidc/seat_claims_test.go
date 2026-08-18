package oidc

import (
	"context"
	"errors"
	"testing"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	seat "github.com/EonsofStupid/tessera/backend/v1/domain"
	"github.com/EonsofStupid/tessera/internal/api/authz"
)

// The rules are tested in `backend/v1/domain` and the SQL in
// `backend/v1/storage/seat`. What is left here is what this adapter itself
// decides: which refusals become which OIDC error, and what happens to a
// request that is not about seats at all.
//
// A fake standing in for the repository is the point of the port. Before it
// existed, testing any of this needed a userinfo read model built by hand.

type fakeSeats struct {
	seat *seat.Seat
	err  error
}

func (f fakeSeats) SeatByMember(context.Context, string, string) (*seat.Seat, error) {
	return f.seat, f.err
}
func (fakeSeats) SetSeat(context.Context, string, *seat.Seat) error { return nil }
func (fakeSeats) RemoveSeat(context.Context, string, string) error  { return nil }
func (fakeSeats) SeatsInWorkspace(context.Context, string, string) ([]*seat.Seat, error) {
	return nil, nil
}

// The instance interceptor is not in play in a unit test, so give the context
// the instance the adapter will ask it for.
func ctxWithInstance() context.Context {
	return authz.WithInstanceID(context.Background(), "inst-1")
}

func serverWith(s *seat.Seat) *Server { return &Server{seats: fakeSeats{seat: s}} }

// Tessera is still an ordinary identity provider for things that are not seats.
// If this regresses, every non-seat integration breaks at once and the error
// looks like it came from somewhere else.
func TestSetSeatClaims_AudienceWithoutAWorkspaceIsLeftAlone(t *testing.T) {
	srv := serverWith(&seat.Seat{MemberID: "mem_1", Workspaces: []string{"ws-0001"}})
	for _, aud := range [][]string{
		{"280895440851832833"},         // a Zitadel project id
		{"280895440851832833@tessera"}, // a client id
		{},                             // none at all
		{"https://some.api.example/"},  // an ordinary resource server
	} {
		claims := &oidc.AccessTokenClaims{}
		if err := srv.setSeatClaims(ctxWithInstance(), claims, aud, "mem_1"); err != nil {
			t.Fatalf("aud=%q: %v — a non-seat token must still mint", aud, err)
		}
		if len(claims.Claims) != 0 {
			t.Errorf("aud=%q: stamped %v onto a token that is not a seat token", aud, claims.Claims)
		}
	}
}

// Every domain refusal has to reach the caller as `invalid_target`, not as
// `server_error` — the second sends an operator to read our logs instead of
// their own configuration.
func TestSetSeatClaims_RefusalsBecomeInvalidTarget(t *testing.T) {
	occupies := &seat.Seat{MemberID: "mem_1", AccountID: "acc", Workspaces: []string{"ws-0001"}}
	unprovisioned := &seat.Seat{MemberID: "mem_1"}

	for name, tc := range map[string]struct {
		seat *seat.Seat
		aud  []string
	}{
		"a workspace the seat does not occupy": {occupies, []string{"automaton:ws-0009"}},
		"an unprovisioned member":              {unprovisioned, []string{"automaton:ws-0001"}},
		"an audience naming two workspaces":    {occupies, []string{"automaton:ws-0001", "devforge:ws-0002"}},
	} {
		claims := &oidc.AccessTokenClaims{}
		err := serverWith(tc.seat).setSeatClaims(ctxWithInstance(), claims, tc.aud, "mem_1")
		var oErr *oidc.Error
		if !errors.As(err, &oErr) || oErr.ErrorType != oidc.InvalidTarget {
			t.Errorf("%s: err = %v, want an *oidc.Error of type invalid_target", name, err)
		}
		if len(claims.Claims) != 0 {
			t.Errorf("%s: claims stamped despite the refusal: %v", name, claims.Claims)
		}
	}
}

// A repository failure is not a refusal. It must not be dressed up as
// `invalid_target`, which would tell an operator their configuration is wrong
// when in fact the database is unreachable.
func TestSetSeatClaims_ALookupFailureIsNotARefusal(t *testing.T) {
	boom := errors.New("connection refused")
	srv := &Server{seats: fakeSeats{err: boom}}
	err := srv.setSeatClaims(ctxWithInstance(), &oidc.AccessTokenClaims{}, []string{"automaton:ws-0001"}, "mem_1")
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want the underlying failure to survive", err)
	}
	var oErr *oidc.Error
	if errors.As(err, &oErr) {
		t.Error("a database failure was reported as an OIDC protocol error")
	}
}

func TestSetSeatClaims_StampsTheContract(t *testing.T) {
	srv := serverWith(&seat.Seat{
		MemberID:      "mem_01J8",
		AccountID:     "acc_01J8",
		Occupant:      seat.OccupantHuman,
		Basis:         seat.BasisSubscription,
		Workspaces:    []string{"ws-0001", "ws-0002"},
		Scopes:        []string{"hosting.active", "terminal:advanced", "chat.unified"},
		PolicyVersion: "pol_2026_08_17",
	})
	claims := &oidc.AccessTokenClaims{}
	if err := srv.setSeatClaims(ctxWithInstance(), claims, []string{"automaton:ws-0001"}, "mem_01J8"); err != nil {
		t.Fatal(err)
	}
	for k, v := range map[string]any{
		"schema":       seat.Schema,
		"account_id":   "acc_01J8",
		"member_id":    "mem_01J8",
		"workspace_id": "ws-0001",
		"occupant":     "human",
		"basis":        "subscription",
	} {
		if claims.Claims[k] != v {
			t.Errorf("claim %q = %v, want %v", k, claims.Claims[k], v)
		}
	}
	authz, ok := claims.Claims["authorization"].(seat.Authorization)
	if !ok {
		t.Fatalf("authorization claim is %T", claims.Claims["authorization"])
	}
	want := []string{"chat:unified", "hosting:active", "terminal:advanced"}
	if len(authz.Scopes) != len(want) {
		t.Fatalf("scopes = %q, want %q", authz.Scopes, want)
	}
	for i := range want {
		if authz.Scopes[i] != want[i] {
			t.Fatalf("scopes = %q, want %q", authz.Scopes, want)
		}
	}
}

// The delegation chain the OIDC layer built is kept, and every actor in it is
// named an agent. `sub` is untouched — that is what makes it delegation.
func TestSetSeatClaims_AnnotatesTheDelegationChain(t *testing.T) {
	claims := &oidc.AccessTokenClaims{
		TokenClaims: oidc.TokenClaims{
			Actor: &oidc.ActorClaims{
				Subject: "clyffy",
				Actor:   &oidc.ActorClaims{Subject: "operator"},
			},
		},
	}
	srv := serverWith(&seat.Seat{
		MemberID:   "mem_01J8",
		AccountID:  "acc",
		Occupant:   seat.OccupantHuman,
		Workspaces: []string{"ws-0001"},
	})
	if err := srv.setSeatClaims(ctxWithInstance(), claims, []string{"automaton:ws-0001"}, "mem_01J8"); err != nil {
		t.Fatal(err)
	}
	if claims.Actor.Subject != "clyffy" || claims.Actor.Claims["occupant"] != "agent" {
		t.Errorf("actor = %+v", claims.Actor)
	}
	if claims.Actor.Actor.Subject != "operator" || claims.Actor.Actor.Claims["occupant"] != "agent" {
		t.Errorf("nested actor = %+v", claims.Actor.Actor)
	}
	if claims.Claims["occupant"] != "human" {
		t.Errorf("the subject's occupant = %v, want human — delegation is not impersonation", claims.Claims["occupant"])
	}
}
