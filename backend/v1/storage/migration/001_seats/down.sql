DROP TRIGGER IF EXISTS trg_set_updated_at_seats ON tessera.seats;
DROP FUNCTION IF EXISTS tessera.set_updated_at();
DROP TABLE IF EXISTS tessera.seat_workspaces;
DROP TABLE IF EXISTS tessera.seats;
DROP TYPE IF EXISTS tessera.seat_basis;
DROP TYPE IF EXISTS tessera.seat_occupant;
