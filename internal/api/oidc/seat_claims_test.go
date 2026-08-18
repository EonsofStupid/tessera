package oidc

import (
	"context"
	"testing"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	seat "github.com/EonsofStupid/tessera/backend/v1/domain"
	seatstorage "github.com/EonsofStupid/tessera/backend/v1/storage/seat"
	"github.com/EonsofStupid/tessera/internal/query"
)

// The rules are tested in `backend/v1/domain` and the fact-reading in
// `backend/v1/storage/seat`. What is left here is what this adapter itself
// decides: which refusals become which OIDC error, and what happens to a
// request that is not about seats at all.

func userInfoWith(orgID string, md map[string]string) *query.OIDCUserInfo {
	qu := &query.OIDCUserInfo{Org: &query.UserInfoOrg{ID: orgID}}
	for k, v := range md {
		qu.Metadata = append(qu.Metadata, query.UserMetadata{Key: k, Value: []byte(v)})
	}
	return qu
}

// Tessera is still an ordinary identity provider for things that are not seats.
// If this regresses, every non-seat integration breaks at once and the error
// looks like it came from somewhere else.
func TestSetSeatClaims_AudienceWithoutAWorkspaceIsLeftAlone(t *testing.T) {
	for _, aud := range [][]string{
		{"280895440851832833"},         // a Zitadel project id
		{"280895440851832833@tessera"}, // a client id
		{},                             // none at all
		{"https://some.api.example/"},  // an ordinary resource server
	} {
		claims := &oidc.AccessTokenClaims{}
		if err := setSeatClaims(context.Background(), claims, userInfoWith("org1", nil), aud, "user1"); err != nil {
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
	occupies := userInfoWith("org1", map[string]string{seatstorage.KeyWorkspaces: "ws-0001"})
	for name, tc := range map[string]struct {
		qu  *query.OIDCUserInfo
		aud []string
	}{
		"a workspace the seat does not occupy": {occupies, []string{"automaton:ws-0009"}},
		"an unprovisioned member":              {userInfoWith("org1", nil), []string{"automaton:ws-0001"}},
		"an audience naming two workspaces":    {occupies, []string{"automaton:ws-0001", "devforge:ws-0002"}},
	} {
		claims := &oidc.AccessTokenClaims{}
		err := setSeatClaims(context.Background(), claims, tc.qu, tc.aud, "mem_1")
		var oErr *oidc.Error
		if !asOIDCError(err, &oErr) || oErr.ErrorType != oidc.InvalidTarget {
			t.Errorf("%s: err = %v, want an *oidc.Error of type invalid_target", name, err)
		}
		if len(claims.Claims) != 0 {
			t.Errorf("%s: claims stamped despite the refusal: %v", name, claims.Claims)
		}
	}
}

func TestSetSeatClaims_StampsTheContract(t *testing.T) {
	claims := &oidc.AccessTokenClaims{}
	qu := userInfoWith("org-tenant-1", map[string]string{
		seatstorage.KeyWorkspaces:    "ws-0001 ws-0002",
		seatstorage.KeyOccupant:      "human",
		seatstorage.KeyBasis:         "subscription",
		seatstorage.KeyScopes:        "hosting.active terminal:advanced chat.unified",
		seatstorage.KeyPolicyVersion: "pol_2026_08_17",
	})
	if err := setSeatClaims(context.Background(), claims, qu, []string{"automaton:ws-0001"}, "mem_01J8"); err != nil {
		t.Fatal(err)
	}
	for k, v := range map[string]any{
		"schema":       seat.Schema,
		"account_id":   "org-tenant-1",
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
	// Dotted spellings arrive from DevForge's contract and leave canonical.
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
	qu := userInfoWith("org1", map[string]string{
		seatstorage.KeyWorkspaces: "ws-0001",
		seatstorage.KeyOccupant:   "human",
	})
	if err := setSeatClaims(context.Background(), claims, qu, []string{"automaton:ws-0001"}, "mem_01J8"); err != nil {
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

func asOIDCError(err error, target **oidc.Error) bool {
	e, ok := err.(*oidc.Error)
	if ok {
		*target = e
	}
	return ok
}
