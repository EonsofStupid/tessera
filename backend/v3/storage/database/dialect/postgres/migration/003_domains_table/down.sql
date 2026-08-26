DROP TABLE IF EXISTS nomen.instance_domains;
DROP TABLE IF EXISTS nomen.org_domains;
DROP TYPE IF EXISTS nomen.domain_type;
DROP TYPE IF EXISTS nomen.domain_validation_type;
DROP FUNCTION IF EXISTS nomen.check_verified_org_domain();
DROP FUNCTION IF EXISTS nomen.ensure_single_primary_instance_domain();
