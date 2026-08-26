// Package seat implements [domain.SeatRepository] over Postgres.
//
// Plain parameterised SQL against a fixed schema rather than v3's statement
// builder. That builder exists to assemble queries from arbitrary caller-supplied
// conditions; seats have four queries and all of them are known at compile time,
// so the DSL would buy nothing and cost readability. It still runs through the
// same [database.Pool], so it shares the connection pool, transactions and
// telemetry with everything else.
package seat

import (
	"context"
	"errors"
	"fmt"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/shippinAI/nomen/backend/v3/storage/database"
)

// Repository reads and writes seats.
type Repository struct {
	pool database.Pool
}

// NewRepository returns a seat repository over the given pool.
func NewRepository(pool database.Pool) *Repository {
	return &Repository{pool: pool}
}

var _ domain.SeatRepository = (*Repository)(nil)

// One round trip on the token path. The workspaces come back as an aggregate
// rather than a second query because this runs on every mint, and the FILTER
// keeps a seat with no workspaces as an empty array instead of `{NULL}`.
const selectSeat = `
SELECT s.account_id, s.occupant::TEXT, s.basis::TEXT, s.scopes, s.policy_version,
       COALESCE(array_agg(w.workspace_id) FILTER (WHERE w.workspace_id IS NOT NULL), '{}')
FROM nomen_product.seats s
LEFT JOIN nomen_product.seat_workspaces w
       ON w.instance_id = s.instance_id AND w.member_id = s.member_id
WHERE s.instance_id = $1 AND s.member_id = $2
GROUP BY s.account_id, s.occupant, s.basis, s.scopes, s.policy_version`

// SeatByMember implements [domain.SeatRepository].
func (r *Repository) SeatByMember(ctx context.Context, instanceID, memberID string) (*domain.Seat, error) {
	seat, _, err := seatByMember(ctx, r.pool, instanceID, memberID)
	return seat, err
}

// seatByMember is the read against whatever executor the caller holds — the
// pool on the token path, the engine's transaction during a blueprint apply.
// It also reports whether a row existed, which the public method deliberately
// hides (an unprovisioned member is a seat that occupies nothing, not an
// error) and the applier needs (created and updated are different outcomes).
func seatByMember(ctx context.Context, qe database.QueryExecutor, instanceID, memberID string) (*domain.Seat, bool, error) {
	seat := &domain.Seat{MemberID: memberID}
	var occupant, basis string

	err := qe.QueryRow(ctx, selectSeat, instanceID, memberID).Scan(
		&seat.AccountID, &occupant, &basis, &seat.Scopes, &seat.PolicyVersion, &seat.Workspaces,
	)
	if err != nil {
		if isNoRows(err) {
			return &domain.Seat{MemberID: memberID}, false, nil
		}
		return nil, false, err
	}

	// Parsed rather than cast. The column is an enum and cannot hold anything
	// else today, but storage is never trusted to have been written by this
	// version of the code — an older or newer value lands on its safe default
	// here rather than on the wire.
	seat.Occupant = domain.ParseOccupant(occupant)
	seat.Basis = domain.ParseBasis(basis)
	return seat, true, nil
}

// SetSeat implements [domain.SeatRepository].
//
// One transaction. The seat row and its workspaces are one fact, and a failure
// between them would leave a seat entitled to workspaces nobody declared — or,
// worse, silently stripped of ones it still has.
func (r *Repository) SetSeat(ctx context.Context, instanceID string, seat *domain.Seat) (err error) {
	if seat == nil || seat.MemberID == "" {
		return errors.New("seat: a seat needs a member")
	}
	if seat.AccountID == "" {
		return fmt.Errorf("seat: %s has no account", seat.MemberID)
	}

	tx, err := r.pool.Begin(ctx, &database.TransactionOptions{
		IsolationLevel: database.IsolationLevelReadCommitted,
		AccessMode:     database.AccessModeReadWrite,
	})
	if err != nil {
		return err
	}
	defer func() { err = tx.End(ctx, err) }()

	return upsertSeat(ctx, tx, instanceID, seat)
}

// upsertSeat is the write against whatever executor the caller holds. SetSeat
// wraps it in its own transaction for callers that arrive without one; the
// blueprint engine passes its own, so the whole file commits or none of it
// does.
func upsertSeat(ctx context.Context, qe database.QueryExecutor, instanceID string, seat *domain.Seat) (err error) {
	// The whole seat, not a patch — a blueprint declares what should be true,
	// so the columns it does not mention must land on their defaults rather
	// than on whatever a previous run left behind.
	if _, err = qe.Exec(ctx, `
INSERT INTO nomen_product.seats (instance_id, member_id, account_id, occupant, basis, scopes, policy_version)
VALUES ($1, $2, $3, $4::nomen_product.seat_occupant, $5::nomen_product.seat_basis, $6, $7)
ON CONFLICT (instance_id, member_id) DO UPDATE SET
    account_id     = EXCLUDED.account_id,
    occupant       = EXCLUDED.occupant,
    basis          = EXCLUDED.basis,
    scopes         = EXCLUDED.scopes,
    policy_version = EXCLUDED.policy_version`,
		instanceID, seat.MemberID, seat.AccountID,
		string(domain.ParseOccupant(string(seat.Occupant))),
		string(domain.ParseBasis(string(seat.Basis))),
		nonNilScopes(seat.Scopes), seat.PolicyVersion,
	); err != nil {
		return err
	}

	// Replaced, not merged. Removing a workspace from a blueprint has to remove
	// the entitlement, or revoking access becomes something only a human can do.
	if _, err = qe.Exec(ctx,
		`DELETE FROM nomen_product.seat_workspaces
		 WHERE instance_id = $1 AND member_id = $2 AND workspace_id <> ALL($3)`,
		instanceID, seat.MemberID, nonNilScopes(seat.Workspaces),
	); err != nil {
		return err
	}
	for _, ws := range seat.Workspaces {
		if _, err = qe.Exec(ctx,
			`INSERT INTO nomen_product.seat_workspaces (instance_id, member_id, workspace_id)
			 VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
			instanceID, seat.MemberID, ws,
		); err != nil {
			return err
		}
	}
	return nil
}

// RemoveSeat implements [domain.SeatRepository]. Workspaces go with it by
// cascade. Removing a seat that is not there is success — `absent` is a desired
// state, and reaching it twice is reaching it.
func (r *Repository) RemoveSeat(ctx context.Context, instanceID, memberID string) error {
	_, err := removeSeat(ctx, r.pool, instanceID, memberID)
	return err
}

// removeSeat reports how many rows went, which the applier turns into the
// difference between `removed` and `unchanged`.
func removeSeat(ctx context.Context, qe database.QueryExecutor, instanceID, memberID string) (int64, error) {
	return qe.Exec(ctx,
		`DELETE FROM nomen_product.seats WHERE instance_id = $1 AND member_id = $2`,
		instanceID, memberID)
}

// SeatsInWorkspace implements [domain.SeatRepository] — the panel's question,
// and the one an array column could not answer without scanning the instance.
func (r *Repository) SeatsInWorkspace(ctx context.Context, instanceID, workspaceID string) ([]*domain.Seat, error) {
	rows, err := r.pool.Query(ctx, `
SELECT s.member_id, s.account_id, s.occupant::TEXT, s.basis::TEXT, s.scopes, s.policy_version
FROM nomen_product.seats s
JOIN nomen_product.seat_workspaces w
  ON w.instance_id = s.instance_id AND w.member_id = s.member_id
WHERE s.instance_id = $1 AND w.workspace_id = $2
ORDER BY s.member_id`, instanceID, workspaceID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	seats := make([]*domain.Seat, 0)
	for rows.Next() {
		s := &domain.Seat{Workspaces: []string{workspaceID}}
		var occupant, basis string
		if err := rows.Scan(&s.MemberID, &s.AccountID, &occupant, &basis, &s.Scopes, &s.PolicyVersion); err != nil {
			return nil, err
		}
		s.Occupant = domain.ParseOccupant(occupant)
		s.Basis = domain.ParseBasis(basis)
		seats = append(seats, s)
	}
	return seats, rows.Err()
}

// nonNilScopes keeps a nil slice from reaching the driver as NULL where the
// column is `NOT NULL DEFAULT '{}'` — an empty set and an absent one mean the
// same thing to us and different things to Postgres.
func nonNilScopes(v []string) []string {
	if v == nil {
		return []string{}
	}
	return v
}

// isNoRows recognises the dialect-independent wrapper the pool returns, rather
// than pgx.ErrNoRows directly — the whole point of the abstraction is that this
// package never learns which driver is underneath.
func isNoRows(err error) bool {
	return errors.Is(err, &database.NoRowFoundError{})
}
