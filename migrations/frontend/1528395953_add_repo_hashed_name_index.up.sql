-- Should run as single non-transaction block
CREATE INDEX CONCURRENTLY IF NOT EXISTS repo_hashed_name_idx ON repo USING BTREE (sha256(lower(name)::bytea));
