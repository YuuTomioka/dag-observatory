#!/usr/bin/env bash
set -euo pipefail

readonly NAME_REGEX='^[0-9]{6}_[a-z0-9_]+\.sql$'
readonly MIGRATE_DIR='data/tsdb/schema/migrate'
readonly SEED_DIR='data/tsdb/schema/seed'

lint_dir() {
  local target_dir="$1"
  local label="$2"
  local files=()
  local invalid=()
  local file=""
  local filename=""

  if [[ ! -d "$target_dir" ]]; then
    echo "$label directory not found: $target_dir" >&2
    return 1
  fi

  shopt -s nullglob
  files=("$target_dir"/*.sql)

  if [[ ${#files[@]} -eq 0 ]]; then
    echo "no SQL files found in $target_dir ($label)"
    return 0
  fi

  for file in "${files[@]}"; do
    filename="$(basename "$file")"
    if [[ ! "$filename" =~ $NAME_REGEX ]]; then
      invalid+=("$filename")
    fi
  done

  if [[ ${#invalid[@]} -gt 0 ]]; then
    echo "invalid $label filenames detected in $target_dir:" >&2
    printf '  - %s\n' "${invalid[@]}" >&2
    echo "expected files to match: NNNNNN_description.sql (lower snake case)" >&2
    return 1
  fi

  echo "$label filename lint passed ($target_dir)"
}

lint_dir "$MIGRATE_DIR" "migration"
lint_dir "$SEED_DIR" "seed"
