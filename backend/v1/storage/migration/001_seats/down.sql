DROP TRIGGER IF EXISTS trg_set_updated_at_seats ON nomen_product.seats;
DROP FUNCTION IF EXISTS nomen_product.set_updated_at();
DROP TABLE IF EXISTS nomen_product.seat_workspaces;
DROP TABLE IF EXISTS nomen_product.seats;
DROP TYPE IF EXISTS nomen_product.seat_basis;
DROP TYPE IF EXISTS nomen_product.seat_occupant;
