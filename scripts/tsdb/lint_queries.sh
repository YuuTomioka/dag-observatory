#!/usr/bin/env bash
set -euo pipefail

readonly QUERY_DIR="data/tsdb/query"
readonly NAME_REGEX='^[0-9]{6}_[a-z0-9_]+\.sql$'

if [[ ! -d "$QUERY_DIR" ]]; then
  echo "query directory not found: $QUERY_DIR" >&2
  exit 1
fi

shopt -s nullglob
files=("$QUERY_DIR"/*.sql)

if [[ ${#files[@]} -eq 0 ]]; then
  echo "no SQL files found in $QUERY_DIR"
  exit 0
fi

invalid=()
for file in "${files[@]}"; do
  filename="$(basename "$file")"
  if [[ ! "$filename" =~ $NAME_REGEX ]]; then
    invalid+=("$filename")
  fi
done

if [[ ${#invalid[@]} -gt 0 ]]; then
  echo "invalid query filenames detected in $QUERY_DIR:" >&2
  printf '  - %s\n' "${invalid[@]}" >&2
  echo "expected files to match: NNNNNN_description.sql (lower snake case)" >&2
  exit 1
fi

echo "query filename lint passed ($QUERY_DIR)"
