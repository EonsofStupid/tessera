// Package seat adapts stored seat facts to the domain's [domain.SeatRepository].
//
// One adapter today, over Zitadel user metadata, because that is where the
// facts currently live. It is the only thing in the tree that knows the
// metadata key names, and that is the point: when blueprints take over
// authoring seats, this file is what gets replaced.
package seat

import (
	"context"
	"strings"

	"github.com/EonsofStupid/tessera/backend/v1/domain"
	"github.com/EonsofStupid/tessera/internal/query"
)

// The keys a seat's facts are stored under, namespaced so they cannot collide
// with metadata a customer sets for their own purposes.
const (
	KeyWorkspaces    = "shippin:seat:workspaces"
	KeyOccupant      = "shippin:seat:occupant"
	KeyBasis         = "shippin:seat:basis"
	KeyAccountID     = "shippin:account_id"
	KeyScopes        = "shippin:entitlement:scopes"
	KeyPolicyVersion = "shippin:entitlement:policy_version"
)

// UserInfoRepository reads a seat out of an already-materialised userinfo read
// model.
//
// It is an adapter and not a query: the OIDC mint path has already fetched this
// row to build the token's other claims, so re-reading it would cost a second
// round trip on the request path to learn what is already in hand. The port is
// still worth having — a blueprint-backed implementation will do its own
// lookup, and this type is what proves the domain never noticed the difference.
type UserInfoRepository struct {
	userInfo *query.OIDCUserInfo
}

// NewUserInfoRepository adapts one userinfo read model. A nil read model is
// valid and yields a seat that occupies nothing.
func NewUserInfoRepository(userInfo *query.OIDCUserInfo) *UserInfoRepository {
	return &UserInfoRepository{userInfo: userInfo}
}

var _ domain.SeatRepository = (*UserInfoRepository)(nil)

// SeatByMember implements [domain.SeatRepository].
func (r *UserInfoRepository) SeatByMember(_ context.Context, memberID string) (*domain.Seat, error) {
	md := r.metadata()

	// The account is the resource-owner organization unless something says
	// otherwise. `account_id` is the tenant today; the key exists so that stays
	// true when organizations grow a level.
	accountID := md[KeyAccountID]
	if accountID == "" && r.userInfo != nil && r.userInfo.Org != nil {
		accountID = r.userInfo.Org.ID
	}

	return &domain.Seat{
		MemberID:  memberID,
		AccountID: accountID,
		Occupant:  domain.ParseOccupant(md[KeyOccupant]),
		Basis:     domain.ParseBasis(md[KeyBasis]),
		// strings.Fields on an absent key yields an empty slice, which is a
		// seat occupying nothing. That is the correct reading of "nobody wrote
		// this down" and it is what makes the gate fail closed.
		Workspaces:    strings.Fields(md[KeyWorkspaces]),
		Scopes:        strings.Fields(md[KeyScopes]),
		PolicyVersion: md[KeyPolicyVersion],
	}, nil
}

// metadata flattens the stored values to plain strings.
//
// These are the values as stored, not the base64 form the
// `urn:zitadel:iam:user:metadata` claim carries. Seat facts are read here and
// re-encoded by the domain; they never travel through that claim, so a client
// which never asked for the metadata scope still gets a correct seat token.
func (r *UserInfoRepository) metadata() map[string]string {
	if r.userInfo == nil || len(r.userInfo.Metadata) == 0 {
		return nil
	}
	out := make(map[string]string, len(r.userInfo.Metadata))
	for _, md := range r.userInfo.Metadata {
		out[md.Key] = string(md.Value)
	}
	return out
}
