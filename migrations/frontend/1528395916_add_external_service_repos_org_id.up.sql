-- +++
-- parent: 1528395915
-- +++

BEGIN;

ALTER TABLE external_service_repos ADD COLUMN IF NOT EXISTS org_id INTEGER REFERENCES orgs(id) ON DELETE CASCADE;

COMMIT;
