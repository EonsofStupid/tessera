package domain

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

// The rule the contract states most emphatically, so it is tested most
// emphatically: a basis nobody measured is not a subscription.
func TestParseBasis_UnknownIsNeverPromoted(t *testing.T) {
	for _, raw := range []string{
		"", "unknown", "none", "nonsense", "SUBSCRIPTION_PENDING",
		"sub", "paid", "trial", "  ", "subscriptionish",
	} {
		if got := ParseBasis(raw); got != BasisUnknown {
			t.Errorf("ParseBasis(%q) = %q, want %q — a basis nobody measured must stay unknown", raw, got, BasisUnknown)
		}
	}
}

func TestParseBasis_KnownSpellings(t *testing.T) {
	cases := map[string]Basis{
		// The canonical spelling.
		"subscription": BasisSubscription,
		"usage":        BasisUsage,
		"local":        BasisLocal,
		// Automaton's internal spelling.
		"subscription_oauth": BasisSubscription,
		"api_key":            BasisUsage,
		// The seam draft's spelling.
		"api": BasisUsage,
		// Case and whitespace are storage accidents, not meaning.
		"  Subscription  ": BasisSubscription,
	}
	for raw, want := range cases {
		if got := ParseBasis(raw); got != want {
			t.Errorf("ParseBasis(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestMint_RefusesAudienceThatNamesNoWorkspace(t *testing.T) {
	// Zitadel puts project and client ids in `aud` as a matter of course.
	// None of them is a tenant boundary.
	for _, aud := range [][]string{
		nil,
		{},
		{"automaton"},
		{"280895440851832833"},
		{"automaton:"},
		{"ws-"},
	} {
		_, err := Mint(Facts{MemberID: "mem_1", Audience: aud})
		if !errors.Is(err, ErrNoWorkspaceAudience) {
			t.Errorf("Mint(aud=%q) err = %v, want ErrNoWorkspaceAudience", aud, err)
		}
	}
}

func TestMint_RefusesAudienceNamingTwoWorkspaces(t *testing.T) {
	_, err := Mint(Facts{
		MemberID: "mem_1",
		Audience: []string{"automaton:ws-0001", "devforge:ws-0002"},
	})
	if !errors.Is(err, ErrAmbiguousWorkspaceAudience) {
		t.Fatalf("err = %v, want ErrAmbiguousWorkspaceAudience — a token audible to two workspaces is the tenant boundary with a hole in it", err)
	}
}

func TestMint_WorkspaceIsDerivedSoItCannotDisagreeWithAud(t *testing.T) {
	// Several services, one workspace, and a project id alongside them — the
	// ordinary shape of a real audience.
	c, err := Mint(Facts{
		MemberID: "mem_01J8",
		Audience: []string{"automaton:ws-0001", "devforge:ws-0001", "280895440851832833"},
	})
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if c.WorkspaceID != "ws-0001" {
		t.Errorf("WorkspaceID = %q, want ws-0001", c.WorkspaceID)
	}
}

func TestMint_DefaultsAreTheSafeOnes(t *testing.T) {
	c, err := Mint(Facts{MemberID: "mem_1", Audience: []string{"automaton:ws-0001"}})
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if c.Basis != BasisUnknown {
		t.Errorf("Basis = %q, want %q for a seat nobody measured", c.Basis, BasisUnknown)
	}
	if c.Occupant != OccupantAgent {
		t.Errorf("Occupant = %q, want %q", c.Occupant, OccupantAgent)
	}
	if c.Schema != Schema {
		t.Errorf("Schema = %q, want %q", c.Schema, Schema)
	}
	// The seam's field and ours carry the same axis and must not drift.
	if c.Provider.AccessClass != c.Basis {
		t.Errorf("provider.access_class %q != basis %q", c.Provider.AccessClass, c.Basis)
	}
}

func TestMint_StoredGarbageDoesNotReachTheWire(t *testing.T) {
	c, err := Mint(Facts{
		MemberID: "mem_1",
		Audience: []string{"automaton:ws-0001"},
		Basis:    Basis("subscription_pending"),
		Occupant: Occupant("robot"),
	})
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if c.Basis != BasisUnknown {
		t.Errorf("Basis = %q, want %q — an unrecognised stored value must land on unknown here, not on the wire", c.Basis, BasisUnknown)
	}
	if c.Occupant != OccupantAgent {
		t.Errorf("Occupant = %q, want %q", c.Occupant, OccupantAgent)
	}
}

func TestMint_RequiresASubject(t *testing.T) {
	if _, err := Mint(Facts{Audience: []string{"automaton:ws-0001"}}); err == nil {
		t.Fatal("want an error for a token with no subject")
	}
}

func TestNormalizeScopes(t *testing.T) {
	got := NormalizeScopes([]string{
		"hosting.active",       // DevForge's dotted spelling
		"terminal:advanced",    // already canonical
		"hosting:active",       // the same entitlement, twice
		"  chat.unified  ",     // whitespace from a config file
		"",                     // nothing
		"urn:zitadel:iam:user", // already has colons; the dot rule must not touch it
	})
	want := []string{"chat:unified", "hosting:active", "terminal:advanced", "urn:zitadel:iam:user"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NormalizeScopes = %q, want %q", got, want)
	}
}

func TestNormalizeScopes_EmptyIsAnEmptyListNotNull(t *testing.T) {
	// `"scopes": null` and `"scopes": []` are different to a consumer that
	// iterates without checking, and the second is the one we mean.
	b, err := json.Marshal(Authorization{Subject: "mem_1", Scopes: NormalizeScopes(nil)})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"subject":"mem_1","scopes":[]}` {
		t.Errorf("marshalled = %s, want scopes as []", got)
	}
}

// The wire format, checked against the contract's own example rather than
// against the struct — a JSON tag typo is exactly the bug this catches.
func TestClaims_WireFormatMatchesTheContract(t *testing.T) {
	c, err := Mint(Facts{
		MemberID:      "mem_01J8",
		AccountID:     "acc_01J8",
		Audience:      []string{"automaton:ws-0001"},
		Occupant:      OccupantHuman,
		Basis:         BasisSubscription,
		Scopes:        []string{"hosting:active", "terminal:advanced", "chat:unified"},
		PolicyVersion: "pol_2026_08_17",
	})
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"schema":       "shippin.seat-token.v1",
		"account_id":   "acc_01J8",
		"member_id":    "mem_01J8",
		"workspace_id": "ws-0001",
		"occupant":     "human",
		"basis":        "subscription",
		"authorization": map[string]any{
			"subject":        "mem_01J8",
			"scopes":         []any{"chat:unified", "hosting:active", "terminal:advanced"},
			"policy_version": "pol_2026_08_17",
		},
		"provider": map[string]any{"access_class": "subscription"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wire format\n got: %#v\nwant: %#v", got, want)
	}
}

// Delegation is RFC 8693's `act`: `sub` stays the human, the agent keeps its
// own identity alongside. An impersonated session has nothing left to indicate,
// which is why §7's visible indicator is only possible under this shape.
func TestClaims_DelegationKeepsTheActorsOwnIdentity(t *testing.T) {
	c, err := Mint(Facts{
		MemberID: "mem_01J8",
		Audience: []string{"automaton:ws-0001"},
		Occupant: OccupantHuman,
		Actor:    &Actor{Subject: "clyffy", Occupant: OccupantAgent},
	})
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	b, _ := json.Marshal(c)
	var got map[string]any
	_ = json.Unmarshal(b, &got)

	if got["sub"] != nil {
		t.Error("`sub` is the OIDC layer's to set, not this package's")
	}
	act, ok := got["act"].(map[string]any)
	if !ok {
		t.Fatalf("no `act` claim in %s", b)
	}
	if act["sub"] != "clyffy" || act["occupant"] != "agent" {
		t.Errorf("act = %#v, want the actor's own identity", act)
	}
	if got["occupant"] != "human" {
		t.Errorf("occupant = %v, want the subject's (human) — delegation is not impersonation", got["occupant"])
	}
}
