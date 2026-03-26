#!/usr/bin/env sh
set -eu

# Minimal migrator runner: applies SQL files in data/tsdb/schema/migrate once.

DB_URL="${TSDB_URL:-${DATABASE_URL:-}}"
if [ -z "$DB_URL" ]; then
  echo "TSDB_URL or DATABASE_URL is required" >&2
  exit 1
fi

MIGRATIONS_DIR="${MIGRATIONS_DIR:-data/tsdb/schema/migrate}"

psql "$DB_URL" -v ON_ERROR_STOP=1 <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
  filename TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
SQL

for file in $(ls -1 "$MIGRATIONS_DIR"/*.sql 2>/dev/null | sort); do
  filename="$(basename "$file")"
  applied="$(psql "$DB_URL" -tA -v ON_ERROR_STOP=1 -c "SELECT 1 FROM schema_migrations WHERE filename='${filename}'")"
  if [ "$applied" = "1" ]; then
    echo "skip $filename"
    continue
  fi
  echo "apply $filename"
  psql "$DB_URL" -v ON_ERROR_STOP=1 -f "$file"
  psql "$DB_URL" -v ON_ERROR_STOP=1 -c "INSERT INTO schema_migrations(filename) VALUES ('${filename}')"
done
