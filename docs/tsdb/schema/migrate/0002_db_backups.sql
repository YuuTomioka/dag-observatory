CREATE TABLE db_backups (
  backup_id       UUID PRIMARY KEY,
  env             TEXT NOT NULL,
  backup_type     TEXT NOT NULL,
  db_name         TEXT NOT NULL,
  schema_version  TEXT NULL,
  bucket          TEXT NOT NULL,
  object_key      TEXT NOT NULL,
  size_bytes      BIGINT NOT NULL CHECK (size_bytes >= 0),
  checksum_sha256 TEXT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  started_at      TIMESTAMPTZ NULL,
  finished_at     TIMESTAMPTZ NULL,
  status          TEXT NOT NULL,
  meta            JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE UNIQUE INDEX ux_db_backups_object
  ON db_backups(bucket, object_key);

CREATE INDEX ix_db_backups_env_created_at
  ON db_backups(env, created_at DESC);
