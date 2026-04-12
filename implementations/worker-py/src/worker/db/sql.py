from __future__ import annotations

from functools import lru_cache
from pathlib import Path
import re


def _find_query_dir() -> Path:
    # Prefer explicit override for deployment environments.
    import os

    override = os.getenv("TSDB_QUERY_DIR")
    if override:
        return Path(override)

    # Walk up from this file to locate data/tsdb/query in repo.
    current = Path(__file__).resolve()
    for parent in current.parents:
        candidate = parent / "data" / "tsdb" / "query"
        if candidate.is_dir():
            return candidate
    raise FileNotFoundError("data/tsdb/query not found; set TSDB_QUERY_DIR")


def _parse_sql_file(path: Path) -> dict[str, str]:
    queries: dict[str, list[str]] = {}
    current_name: str | None = None
    for line in path.read_text(encoding="utf-8").splitlines(keepends=True):
        if line.startswith("-- name:"):
            parts = line.strip().split()
            if len(parts) >= 3:
                current_name = parts[2]
                queries[current_name] = []
            continue
        if current_name:
            queries[current_name].append(line)
    return {name: _normalize_placeholders("".join(lines).strip()) for name, lines in queries.items()}


def _normalize_placeholders(query: str) -> str:
    # Shared SQL assets use PostgreSQL-style placeholders ($1, $2, ...).
    # psycopg expects %s style parameters for execute(..., params).
    return re.sub(r"\$\d+", "%s", query)


@lru_cache
def load_queries() -> dict[str, str]:
    query_dir = _find_query_dir()
    queries: dict[str, str] = {}
    for file in sorted(query_dir.glob("*.sql")):
        queries.update(_parse_sql_file(file))
    return queries


def get_query(name: str) -> str:
    queries = load_queries()
    if name not in queries:
        raise KeyError(f"query not found: {name}")
    return queries[name]
