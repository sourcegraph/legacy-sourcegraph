BEGIN;

ALTER TABLE campaigns ALTER COLUMN description DROP NOT NULL;

COMMIT;
