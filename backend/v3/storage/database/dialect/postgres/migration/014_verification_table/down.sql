ALTER TABLE nomen.users DROP CONSTRAINT IF EXISTS fk_unverified_password;
ALTER TABLE nomen.users DROP CONSTRAINT IF EXISTS fk_unverified_email;
ALTER TABLE nomen.users DROP CONSTRAINT IF EXISTS fk_unverified_phone;

DROP TRIGGER IF EXISTS user_verification_integrity_trigger ON nomen.users;
DROP FUNCTION IF EXISTS nomen.ensure_user_verification_integrity() CASCADE;
DROP TABLE IF EXISTS nomen.verifications CASCADE;
