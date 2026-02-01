#!/usr/bin/env bash
set -euo pipefail

SPEC_PATH="${SPEC_PATH:-openapi/dag-core/openapi.yaml}"
BASE_REF="${BASE_REF:-origin/main}"

tmp_spec="$(mktemp)"
cleanup() {
  rm -f "${tmp_spec}"
}
trap cleanup EXIT

git show "${BASE_REF}:${SPEC_PATH}" > "${tmp_spec}"
go run github.com/oasdiff/oasdiff@latest breaking "${tmp_spec}" "${SPEC_PATH}"
