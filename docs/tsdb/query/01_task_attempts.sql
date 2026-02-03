-- name: UpsertTaskAttemptStarted :exec
INSERT INTO task_attempts (
  task_id,
  attempt_no,
  workflow_run_id,
  status,
  started_at,
  worker_id,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, now()
)
ON CONFLICT (task_id, attempt_no) DO UPDATE
SET status = EXCLUDED.status,
    started_at = COALESCE(task_attempts.started_at, EXCLUDED.started_at),
    worker_id = COALESCE(task_attempts.worker_id, EXCLUDED.worker_id),
    updated_at = now();

-- name: UpdateTaskAttemptCompleted :exec
UPDATE task_attempts
SET status = $3,
    finished_at = $4,
    output_ref = $5,
    updated_at = now()
WHERE task_id = $1 AND attempt_no = $2;

-- name: UpdateTaskAttemptFailed :exec
UPDATE task_attempts
SET status = $3,
    finished_at = $4,
    error = $5,
    updated_at = now()
WHERE task_id = $1 AND attempt_no = $2;
