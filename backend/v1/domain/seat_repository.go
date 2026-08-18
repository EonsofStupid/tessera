package domain

import "context"

// SeatRepository is where a seat's stored facts come from.
//
// The port exists so the domain never learns where seats live. Today they live
// in Zitadel user metadata, which is where they are *stored* and not where they
// should be *authored*; blueprints are what will write them. When that lands it
// replaces the adapter and nothing on this side of the interface changes —
// which is the entire reason for declaring it before there are two
// implementations rather than after.
//
// It returns a [Seat] and not a token. Minting is [Seat.Token], and keeping the
// two apart is what stops "which workspaces may this member occupy" from being
// answered differently by each caller that needs a token.
type SeatRepository interface {
	// SeatByMember returns the seat for a member.
	//
	// A member with no seat facts recorded is not an error — it is a seat that
	// occupies nothing, and [Seat.Token] will refuse every workspace it is
	// asked for. Returning an error here instead would make an unprovisioned
	// member indistinguishable from a broken lookup, and those need different
	// fixes.
	SeatByMember(ctx context.Context, memberID string) (*Seat, error)
}
