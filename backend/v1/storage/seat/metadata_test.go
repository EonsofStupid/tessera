package seat

import (
	"context"
	"reflect"
	"testing"

	"github.com/EonsofStupid/tessera/backend/v1/domain"
	"github.com/EonsofStupid/tessera/internal/query"
)

func userInfo(orgID string, md map[string]string) *query.OIDCUserInfo {
	qu := &query.OIDCUserInfo{Org: &query.UserInfoOrg{ID: orgID}}
	for k, v := range md {
		qu.Metadata = append(qu.Metadata, query.UserMetadata{Key: k, Value: []byte(v)})
	}
	return qu
}

func TestSeatByMember_ReadsTheStoredFacts(t *testing.T) {
	repo := NewUserInfoRepository(userInfo("org-tenant-1", map[string]string{
		KeyWorkspaces:    "ws-0001 ws-0002",
		KeyOccupant:      "human",
		KeyBasis:         "subscription",
		KeyScopes:        "hosting.active terminal:advanced chat.unified",
		KeyPolicyVersion: "pol_2026_08_17",
	}))
	got, err := repo.SeatByMember(context.Background(), "mem_01J8")
	if err != nil {
		t.Fatal(err)
	}
	want := &domain.Seat{
		MemberID:      "mem_01J8",
		AccountID:     "org-tenant-1", // the resource-owner org, since no override
		Occupant:      domain.OccupantHuman,
		Basis:         domain.BasisSubscription,
		Workspaces:    []string{"ws-0001", "ws-0002"},
		Scopes:        []string{"hosting.active", "terminal:advanced", "chat.unified"},
		PolicyVersion: "pol_2026_08_17",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\n got: %+v\nwant: %+v", got, want)
	}
}

// A member nobody provisioned is a seat that occupies nothing — not an error.
// The two need different fixes and an error here would make them look alike.
func TestSeatByMember_NoMetadataIsAnEmptySeatNotAnError(t *testing.T) {
	for _, qu := range []*query.OIDCUserInfo{nil, userInfo("org1", nil)} {
		got, err := NewUserInfoRepository(qu).SeatByMember(context.Background(), "mem_1")
		if err != nil {
			t.Fatalf("%v", err)
		}
		if len(got.Workspaces) != 0 {
			t.Errorf("workspaces = %v, want none", got.Workspaces)
		}
		if got.Basis != domain.BasisUnknown {
			t.Errorf("basis = %q, want unknown", got.Basis)
		}
		if got.Occupant != domain.OccupantAgent {
			t.Errorf("occupant = %q, want agent", got.Occupant)
		}
	}
}

func TestSeatByMember_AccountIDOverridesTheOrg(t *testing.T) {
	repo := NewUserInfoRepository(userInfo("org1", map[string]string{KeyAccountID: "acc_01J8"}))
	got, _ := repo.SeatByMember(context.Background(), "mem_1")
	if got.AccountID != "acc_01J8" {
		t.Errorf("account_id = %q, want the override", got.AccountID)
	}
}

// Storage may hold a value this version has never heard of. It must not arrive
// at the domain looking like something meaningful.
func TestSeatByMember_UnrecognisedValuesLandOnTheirSafeDefault(t *testing.T) {
	repo := NewUserInfoRepository(userInfo("org1", map[string]string{
		KeyBasis:    "subscription_pending",
		KeyOccupant: "robot",
	}))
	got, _ := repo.SeatByMember(context.Background(), "mem_1")
	if got.Basis != domain.BasisUnknown {
		t.Errorf("basis = %q, want unknown", got.Basis)
	}
	if got.Occupant != domain.OccupantAgent {
		t.Errorf("occupant = %q, want agent", got.Occupant)
	}
}
