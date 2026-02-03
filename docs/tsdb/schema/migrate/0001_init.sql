-- Core tables (minimal set)

CREATE TABLE task_attempts (
  task_id        TEXT NOT NULL,
  attempt_no     INT  NOT NULL,
  workflow_run_id TEXT NULL,
  status         TEXT NOT NULL,
  started_at     TIMESTAMPTZ NULL,
  finished_at    TIMESTAMPTZ NULL,
  worker_id      TEXT NULL,
  error          TEXT NULL,
  output_ref     TEXT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (task_id, attempt_no)
);

CREATE TABLE processed_events (
  event_id    TEXT PRIMARY KEY,
  task_id     TEXT NULL,
  attempt_no  INT  NULL,
  received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE outbox_events (
  id              BIGSERIAL PRIMARY KEY,
  event_type      TEXT NOT NULL,
  partition_key   TEXT NOT NULL,
  payload_json    JSONB NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  published_at    TIMESTAMPTZ NULL,
  attempts        INT NOT NULL DEFAULT 0,
  next_publish_at TIMESTAMPTZ NULL,
  last_error      TEXT NULL
);

CREATE INDEX ix_outbox_events_pending
  ON outbox_events(created_at)
  WHERE published_at IS NULL;

CREATE TABLE artifacts (
  artifact_id     UUID PRIMARY KEY,
  workflow_run_id TEXT NOT NULL,
  task_id         TEXT NOT NULL,
  attempt_no      INT  NOT NULL,
  bucket          TEXT NOT NULL,
  object_key      TEXT NOT NULL,
  kind            TEXT NOT NULL,
  filename        TEXT NULL,
  content_type    TEXT NULL,
  size_bytes      BIGINT NOT NULL CHECK (size_bytes >= 0),
  checksum_sha256 TEXT NULL,
  etag            TEXT NULL,
  status          TEXT NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  tags            JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE UNIQUE INDEX ux_artifacts_object
  ON artifacts(bucket, object_key);

CREATE INDEX ix_artifacts_task
  ON artifacts(workflow_run_id, task_id, attempt_no);

CREATE INDEX ix_artifacts_created_at
  ON artifacts(created_at);
