CREATE OR REPLACE TRIGGER trigger_set_updated_at
BEFORE UPDATE ON nomen.instances
FOR EACH ROW
WHEN (NEW.updated_at IS NULL)
EXECUTE FUNCTION nomen.set_updated_at();

CREATE OR REPLACE TRIGGER trigger_set_updated_at
BEFORE UPDATE ON nomen.organizations
FOR EACH ROW
WHEN (NEW.updated_at IS NULL)
EXECUTE FUNCTION nomen.set_updated_at();

CREATE OR REPLACE TRIGGER trg_set_updated_at_instance_domains
  BEFORE UPDATE ON nomen.instance_domains
  FOR EACH ROW
  WHEN (NEW.updated_at IS NULL)
  EXECUTE FUNCTION nomen.set_updated_at();

CREATE OR REPLACE TRIGGER trg_set_updated_at_org_domains
  BEFORE UPDATE ON nomen.org_domains
  FOR EACH ROW
  WHEN (NEW.updated_at IS NULL)
  EXECUTE FUNCTION nomen.set_updated_at();