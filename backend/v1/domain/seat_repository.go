package domain

import "context"

// SeatRepository is where seats live.
//
// The port exists so the domain never learns where that is. It returns a [Seat]
// and never a token — minting is [Seat.Token], and keeping the two apart is what
// stops "which workspaces may this member occupy" from being answered
// differently by each caller that happens to need a token.
type SeatRepository interface {
	// SeatByMember returns the seat for a member.
	//
	// A member with no seat is not an error — it is a seat that occupies
	// nothing, and [Seat.Token] will refuse every workspace it is asked for.
	// Returning an error would make an unprovisioned member indistinguishable
	// from a broken lookup, and those need different fixes.
	SeatByMember(ctx context.Context, instanceID, memberID string) (*Seat, error)

	// SetSeat writes a seat and the exact set of workspaces it occupies,
	// creating it if absent.
	//
	// The whole seat, not a patch. A blueprint declares what should be true,
	// so applying one twice has to land in the same place — and a partial
	// update leaves whatever the previous run wrote for the fields this one
	// omitted, which is how two blueprints silently disagree about a seat
	// neither of them fully describes.
	//
	// Workspaces are replaced rather than merged for the same reason: removing
	// a workspace from a blueprint has to remove the entitlement, or revoking
	// access becomes something you can only do by hand.
	SetSeat(ctx context.Context, instanceID string, seat *Seat) error

	// RemoveSeat deletes a seat. Removing one that is not there is not an
	// error — `absent` is a desired state, and reaching it twice is success
	// both times.
	RemoveSeat(ctx context.Context, instanceID, memberID string) error

	// SeatsInWorkspace lists the seats occupying a workspace.
	//
	// The reverse of the token path, and the question the panel asks: who is in
	// this workspace. It exists on the port because an operator needing it
	// would otherwise reach around the repository and write the SQL somewhere
	// that has no business knowing the schema.
	SeatsInWorkspace(ctx context.Context, instanceID, workspaceID string) ([]*Seat, error)
}
