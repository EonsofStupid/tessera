DROP TABLE IF EXISTS nomen.user_identity_provider_links;
DROP TABLE IF EXISTS nomen.user_passkeys;
DROP TYPE IF EXISTS nomen.passkey_type;
DROP TABLE IF EXISTS nomen.machine_keys;
DROP TABLE IF EXISTS nomen.user_personal_access_tokens;
DROP TABLE IF EXISTS nomen.user_metadata CASCADE;

ALTER TABLE nomen.identity_provider_intents DROP CONSTRAINT IF EXISTS fk_idp_intent_user;
ALTER TABLE nomen.sessions DROP CONSTRAINT IF EXISTS fk_session_user;
ALTER TABLE nomen.authorizations DROP CONSTRAINT IF EXISTS fk_authorization_user;

DROP TABLE IF EXISTS nomen.users;
DROP TYPE IF EXISTS nomen.user_type;
DROP TYPE IF EXISTS nomen.user_state;
