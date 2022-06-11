ALTER TABLE batch_spec_workspace_execution_jobs ADD COLUMN IF NOT EXISTS user_id INTEGER;

UPDATE batch_spec_workspace_execution_jobs exec SET user_id = (
    SELECT spec.user_id
    FROM batch_spec_workspaces AS workspace
    JOIN batch_specs spec ON spec.id = workspace.batch_spec_id
    WHERE workspace.id = exec.batch_spec_workspace_id
);

CREATE INDEX IF NOT EXISTS batch_spec_workspace_execution_jobs_user_id ON batch_spec_workspace_execution_jobs (user_id);

CREATE INDEX IF NOT EXISTS batch_spec_workspace_execution_jobs_state ON batch_spec_workspace_execution_jobs (state);

DROP VIEW IF EXISTS batch_spec_workspace_execution_jobs_with_rank;
DROP VIEW IF EXISTS batch_spec_workspace_execution_queue;

CREATE VIEW batch_spec_workspace_execution_queue AS
WITH user_queues AS (
    SELECT
        exec.user_id,
        MAX(exec.started_at) AS latest_dequeue
    FROM batch_spec_workspace_execution_jobs AS exec
    GROUP BY exec.user_id
),
queue_candidates AS (
    SELECT
        exec.id,
        RANK() OVER (
            PARTITION BY queue.user_id
            -- Make sure the jobs are still fulfilled in timely order, and that the ordering is stable.
            ORDER BY exec.created_at ASC, exec.id ASC
        ) AS place_in_user_queue
    FROM batch_spec_workspace_execution_jobs exec
    JOIN user_queues queue ON queue.user_id = exec.user_id
    WHERE
    	-- Only queued records should get a rank.
        exec.state = 'queued'
    ORDER BY
        -- Round-robin let users dequeue jobs.
        place_in_user_queue,
        -- And ensure the user who dequeued the longest ago is next.
        queue.latest_dequeue ASC NULLS FIRST
)
SELECT
    queue_candidates.id, ROW_NUMBER() OVER () AS place_in_global_queue, queue_candidates.place_in_user_queue
FROM queue_candidates;

CREATE VIEW batch_spec_workspace_execution_jobs_with_rank AS (
    SELECT
        j.*,
        q.place_in_global_queue,
        q.place_in_user_queue
    FROM
        batch_spec_workspace_execution_jobs j
    LEFT JOIN batch_spec_workspace_execution_queue q ON j.id = q.id
);
