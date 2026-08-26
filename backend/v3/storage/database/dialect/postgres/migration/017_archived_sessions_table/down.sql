DROP TRIGGER IF EXISTS trg_move_to_archived_sessions ON nomen.sessions;
DROP FUNCTION IF EXISTS nomen.move_to_archived_sessions() CASCADE;
DROP TABLE IF EXISTS nomen.archived_sessions CASCADE;
DROP FUNCTION IF EXISTS nomen.throw_not_permitted();