BEGIN;

ALTER TABLE campaign_plans ADD COLUMN user_id integer;

-- Set to first site admin if not attached
UPDATE campaign_plans SET user_id = users.id
FROM users
WHERE site_admin;

-- Use campaign.author_id for campaign_plans that are already
-- attached to a campaign
UPDATE campaign_plans SET user_id = author_id
FROM campaigns
WHERE campaign_plan_id = campaign_plans.id;

ALTER TABLE campaign_plans
ADD CONSTRAINT campaign_plans_user_id_fkey FOREIGN KEY (user_id)
REFERENCES users (id) DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE campaign_plans ALTER COLUMN user_id SET NOT NULL;

COMMIT;
