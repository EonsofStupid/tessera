package oidc

import (
	"errors"
	"testing"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	seat "github.com/EonsofStupid/tessera/backend/v1/domain"
	"github.com/EonsofStupid/tessera/internal/query"
)

func userInfoWith(orgID string, md map[string]string) *query.OIDCUserInfo {
	qu := &query.OIDCUserInfo{Org: &query.UserInfoOrg{ID: orgID}}
	for k, v := range md {
		qu.Metadata = append(qu.Metadata, query.UserMetadata{Key: k, Value: []byte(v)})
	}
	return qu
}

// Tessera is still an ordinary identity provider for things that are not
// seats. A client asking for an ordinary token must get one, unchanged — if
// this regresses, every non-seat integration breaks at once and the error will
// look like it came from somewhere else.
func TestSetSeatClaims_AudienceWithoutAWorkspaceIsLeftAlone(t *testing.T) {
	for _, aud := range [][]string{
		{"280895440851832833"},         // a Zitadel project id
		{"280895440851832833@tessera"}, // a client id
		{},                             // none at all
		{"https://some.api.example/"},  // an ordinary resource server
	} {
		claims := &oidc.AccessTokenClaims{}
		if err := setSeatClaims(claims, userInfoWith("org1", nil), aud, "user1"); err != nil {
			t.Fatalf("aud=%q: %v — a non-seat token must still mint", aud, err)
		}
		if len(claims.Claims) != 0 {
			t.Errorf("aud=%q: stamped %v onto a token that is not a seat token", aud, claims.Claims)
		}
	}
}

// The one case that must fail closed.
func TestSetSeatClaims_RefusesAnAmbiguousWorkspace(t *testing.T) {
	claims := &oidc.AccessTokenClaims{}
	err := setSeatClaims(claims, userInfoWith("org1", nil), []string{"automaton:ws-0001", "devforge:ws-0002"}, "user1")
	if err == nil {
		t.Fatal("want a refusal: a token audible to two workspaces has no tenant boundary")
	}
	if len(claims.Claims) != 0 {
		t.Errorf("claims were stamped despite the refusal: %v", claims.Claims)
	}
}

func TestSetSeatClaims_StampsTheContract(t *testing.T) {
	claims := &oidc.AccessTokenClaims{}
	qu := userInfoWith("org-tenant-1", map[string]string{
		metadataKeyWorkspaces:    "ws-0001 ws-0002",
		metadataKeyOccupant:      "human",
		metadataKeyBasis:         "subscription",
		metadataKeyScopes:        "hosting.active terminal:advanced chat.unified",
		metadataKeyPolicyVersion: "pol_2026_08_17",
	})
	if err := setSeatClaims(claims, qu, []string{"automaton:ws-0001"}, "mem_01J8"); err != nil {
		t.Fatal(err)
	}

	want := map[string]any{
		"schema":       seat.Schema,
		"account_id":   "org-tenant-1", // the resource-owner org, since no override
		"member_id":    "mem_01J8",
		"workspace_id": "ws-0001",
		"occupant":     "human",
		"basis":        "subscription",
	}
	for k, v := range want {
		if claims.Claims[k] != v {
			t.Errorf("claim %q = %v, want %v", k, claims.Claims[k], v)
		}
	}

	authz, ok := claims.Claims["authorization"].(seat.Authorization)
	if !ok {
		t.Fatalf("authorization claim is %T", claims.Claims["authorization"])
	}
	// Dotted spellings arrive from DevForge's contract and leave in the
	// canonical colon form, sorted.
	wantScopes := []string{"chat:unified", "hosting:active", "terminal:advanced"}
	if len(authz.Scopes) != len(wantScopes) {
		t.Fatalf("scopes = %q, want %q", authz.Scopes, wantScopes)
	}
	for i, s := range wantScopes {
		if authz.Scopes[i] != s {
			t.Errorf("scopes = %q, want %q", authz.Scopes, wantScopes)
			break
		}
	}
	if authz.PolicyVersion != "pol_2026_08_17" {
		t.Errorf("policy_version = %q", authz.PolicyVersion)
	}

	prov, ok := claims.Claims["provider"].(seat.Provider)
	if !ok || prov.AccessClass != seat.BasisSubscription {
		t.Errorf("provider = %v", claims.Claims["provider"])
	}
}

// A seat whose facts were never written is a seat nobody measured, and the
// token has to say so rather than omit it.
func TestSetSeatClaims_NoMetadataIsUnknownNotAbsent(t *testing.T) {
	claims := &oidc.AccessTokenClaims{}
	qu := userInfoWith("org1", map[string]string{metadataKeyWorkspaces: "ws-0001"})
	if err := setSeatClaims(claims, qu, []string{"automaton:ws-0001"}, "mem_1"); err != nil {
		t.Fatal(err)
	}
	if claims.Claims["basis"] != "unknown" {
		t.Errorf("basis = %v, want unknown", claims.Claims["basis"])
	}
	if claims.Claims["occupant"] != "agent" {
		t.Errorf("occupant = %v, want agent", claims.Claims["occupant"])
	}
}

// Metadata that says "subscription" in a spelling nobody defined must not
// become a subscription just because it starts with the right letters.
func TestSetSeatClaims_UnknownIsNeverPromotedThroughMetadata(t *testing.T) {
	claims := &oidc.AccessTokenClaims{}
	qu := userInfoWith("org1", map[string]string{
		metadataKeyWorkspaces: "ws-0001",
		metadataKeyBasis:      "subscription_pending",
	})
	if err := setSeatClaims(claims, qu, []string{"automaton:ws-0001"}, "mem_1"); err != nil {
		t.Fatal(err)
	}
	if claims.Claims["basis"] != "unknown" {
		t.Errorf("basis = %v, want unknown", claims.Claims["basis"])
	}
	if claims.Claims["provider"].(seat.Provider).AccessClass != seat.BasisUnknown {
		t.Error("provider.access_class drifted from basis")
	}
}

func TestSetSeatClaims_AccountIDOverridesTheOrg(t *testing.T) {
	claims := &oidc.AccessTokenClaims{}
	qu := userInfoWith("org1", map[string]string{
		metadataKeyWorkspaces: "ws-0001",
		metadataKeyAccountID:  "acc_01J8",
	})
	if err := setSeatClaims(claims, qu, []string{"automaton:ws-0001"}, "mem_1"); err != nil {
		t.Fatal(err)
	}
	if claims.Claims["account_id"] != "acc_01J8" {
		t.Errorf("account_id = %v, want the override", claims.Claims["account_id"])
	}
}

// Delegation: the chain the OIDC layer built is kept, and every actor in it is
// named as an agent. `sub` is untouched — that is what makes it delegation
// rather than impersonation.
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
		metadataKeyWorkspaces: "ws-0001",
		metadataKeyOccupant:   "human",
	})
	if err := setSeatClaims(claims, qu, []string{"automaton:ws-0001"}, "mem_01J8"); err != nil {
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

// The gate: naming a workspace in a scope is a request, not a permission.
func TestSetSeatClaims_RefusesAWorkspaceTheMemberDoesNotOccupy(t *testing.T) {
	claims := &oidc.AccessTokenClaims{}
	qu := userInfoWith("org1", map[string]string{metadataKeyWorkspaces: "ws-0001 ws-0002"})
	err := setSeatClaims(claims, qu, []string{"automaton:ws-0009"}, "mem_1")
	if !errors.Is(err, ErrWorkspaceNotOccupied) {
		t.Fatalf("err = %v, want ErrWorkspaceNotOccupied", err)
	}
	if len(claims.Claims) != 0 {
		t.Errorf("claims stamped despite the refusal: %v", claims.Claims)
	}
}

// An unprovisioned member is not a member with universal access. This is the
// one that would fail open if the check were written as "no list means no
// restriction", which is the tempting way to write it.
func TestSetSeatClaims_NoRecordedWorkspacesMeansNone(t *testing.T) {
	claims := &oidc.AccessTokenClaims{}
	err := setSeatClaims(claims, userInfoWith("org1", nil), []string{"automaton:ws-0001"}, "mem_1")
	if !errors.Is(err, ErrWorkspaceNotOccupied) {
		t.Fatalf("err = %v, want ErrWorkspaceNotOccupied — an unprovisioned member occupies nothing", err)
	}
}

// A member occupying ws-0001 must not reach ws-00010 by prefix, nor ws-0001 by
// naming a substring of it.
func TestSetSeatClaims_WorkspaceMatchIsExact(t *testing.T) {
	qu := userInfoWith("org1", map[string]string{metadataKeyWorkspaces: "ws-0001"})
	for _, ws := range []string{"ws-00010", "ws-000", "ws-0001x"} {
		claims := &oidc.AccessTokenClaims{}
		if err := setSeatClaims(claims, qu, []string{"automaton:" + ws}, "mem_1"); !errors.Is(err, ErrWorkspaceNotOccupied) {
			t.Errorf("%s: err = %v, want ErrWorkspaceNotOccupied", ws, err)
		}
	}
}
