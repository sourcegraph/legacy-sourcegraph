BEGIN;

CREATE TABLE IF NOT EXISTS insights_settings_migration_jobs
(
    id SERIAL NOT NULL,
    user_id int,
    org_id int,
    global boolean,
    settings_id int NOT NULL, -- non-constrained foreign key to settings object that should be migrated
    total_insights int NOT NULL DEFAULT 0,
    migrated_insights int NOT NULL DEFAULT 0,
    total_dashboards int NOT NULL DEFAULT 0,
    migrated_dashboards int NOT NULL DEFAULT 0,
    virtual_dashboard_created_at TIMESTAMP,
    runs int NOT NULL DEFAULT 0,
    error_msg TEXT,
    completed_at timestamp
);

-- We go in this order (global, org, user) such that we migrate any higher level shared insights first. This way
-- we can just go in the order of id rather than have a secondary index.

-- global
INSERT INTO insights_settings_migration_jobs (global, settings_id)
SELECT TRUE, MAX(id)
FROM settings
WHERE user_id IS NULL
  AND org_id IS NULL;

-- org
INSERT INTO insights_settings_migration_jobs (settings_id, org_id)
SELECT DISTINCT ON (org_id) id, org_id
FROM settings
WHERE org_id IS NOT NULL
ORDER BY org_id, id DESC;

--  user
INSERT INTO insights_settings_migration_jobs (settings_id, user_id)
SELECT DISTINCT ON (user_id) id, user_id
FROM settings
WHERE user_id IS NOT NULL
ORDER BY user_id, id DESC;

COMMIT;
