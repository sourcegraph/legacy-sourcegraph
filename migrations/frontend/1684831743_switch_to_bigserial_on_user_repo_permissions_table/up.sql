LOCK user_repo_permissions IN EXCLUSIVE MODE;

-- drop primary key constraint first
ALTER TABLE user_repo_permissions 
DROP CONSTRAINT IF EXISTS user_repo_permissions_pkey;

-- change the id column to bigint
ALTER TABLE user_repo_permissions 
ALTER COLUMN id TYPE BIGINT;

-- update the sequence
ALTER SEQUENCE user_repo_permissions_id_seq AS BIGINT OWNED BY user_repo_permissions.id RESTART WITH 1;

-- reassign all the primary keys
UPDATE user_repo_permissions 
SET id = nextval('user_repo_permissions_id_seq');

-- add back the primary key constraint
ALTER TABLE user_repo_permissions ADD CONSTRAINT user_repo_permissions_pkey PRIMARY KEY (id);
