-- Seats: a licensed occupant of one or more workspaces.
--
-- Until now these lived in Nomen user metadata, which is a key-value bag with
-- no types, no constraints and no way to ask "who occupies ws-0001". That is
-- survivable for one seat and indefensible for a fleet, which is what this
-- table is for.

CREATE TYPE nomen_product.seat_occupant AS ENUM (
    'human'
    , 'agent'
);

-- `unknown` is a real value and not a null. A basis nobody measured is not a
-- subscription, and modelling it as NULL invites a COALESCE somewhere that
-- turns "not measured" into "the cheapest thing that parses".
CREATE TYPE nomen_product.seat_basis AS ENUM (
    'subscription'
    , 'usage'
    , 'local'
    , 'unknown'
);

CREATE TABLE nomen_product.seats (
    instance_id TEXT NOT NULL
    , member_id TEXT NOT NULL

    , account_id TEXT NOT NULL CHECK (account_id <> '')

    -- Defaults are the safe ones, stated twice on purpose: the domain refuses
    -- to promote an unrecognised value, and the column will not store one.
    -- Defence in depth is cheap here and the failure it prevents is a customer
    -- being billed for capacity nobody chose.
    , occupant nomen_product.seat_occupant NOT NULL DEFAULT 'agent'
    , basis nomen_product.seat_basis NOT NULL DEFAULT 'unknown'

    -- Scopes are an array rather than a child table because they are only ever
    -- read *with* the seat and never searched across seats. Every token mint
    -- reads this row; a join on the hot path to reassemble a set nobody queries
    -- independently would be cost for nothing. Workspaces below are the
    -- opposite case, and are modelled the opposite way.
    , scopes TEXT[] NOT NULL DEFAULT '{}'
    , policy_version TEXT NOT NULL DEFAULT ''

    , created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    , updated_at TIMESTAMPTZ NOT NULL DEFAULT now()

    , PRIMARY KEY (instance_id, member_id)

    -- A seat is about a member. When the member goes, the seat goes with it —
    -- an orphaned seat is an entitlement with nobody attached, and it would
    -- keep answering "yes" to a question nobody should still be asking.
    , CONSTRAINT fk_seat_member FOREIGN KEY (instance_id, member_id)
        REFERENCES nomen.users (instance_id, id) ON DELETE CASCADE
);

-- Which seats belong to an account: the panel's list view, per customer.
CREATE INDEX idx_seats_account ON nomen_product.seats (instance_id, account_id);

-- Workspaces are a child table because this relation is read in *both*
-- directions and only one of them is the token path:
--
--   seat -> workspaces   on every mint, to answer "may this seat have ws-X"
--   workspace -> seats   for the panel and for fleet operations, to answer
--                        "who occupies ws-X" — which an array column cannot
--                        answer without scanning every seat in the instance.
CREATE TABLE nomen_product.seat_workspaces (
    instance_id TEXT NOT NULL
    , member_id TEXT NOT NULL
    , workspace_id TEXT NOT NULL CHECK (workspace_id <> '')

    , created_at TIMESTAMPTZ NOT NULL DEFAULT now()

    , PRIMARY KEY (instance_id, member_id, workspace_id)

    , CONSTRAINT fk_seat_workspace_seat FOREIGN KEY (instance_id, member_id)
        REFERENCES nomen_product.seats (instance_id, member_id) ON DELETE CASCADE
);

-- The reverse lookup. Without it "who occupies ws-0001" is a sequential scan,
-- and that question gets asked once per workspace per panel page load.
CREATE INDEX idx_seat_workspaces_workspace
    ON nomen_product.seat_workspaces (instance_id, workspace_id);

-- Our own, rather than calling nomen.set_updated_at across the schema
-- boundary. Five lines to keep this schema standing on its own.
CREATE OR REPLACE FUNCTION nomen_product.set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER trg_set_updated_at_seats
    BEFORE UPDATE ON nomen_product.seats
    FOR EACH ROW
    EXECUTE FUNCTION nomen_product.set_updated_at();
