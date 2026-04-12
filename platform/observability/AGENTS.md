# AGENTS

## Purpose

This directory is the canonical home for observability platform topology and backend reference configuration.

Repository-wide structure is defined above this directory. Read root `AGENTS.md` and relevant `blueprint/` documents first.

## Local Read Order

1. `README.md`
2. root `blueprint/architecture/system-boundaries.md`
3. root `blueprint/conventions/observability-conventions.md`
4. then the target asset under `compose/`, `otel-collector/`, `grafana/`, `loki/`, `prometheus/`, `promtail/`, or `tempo/`

## Local Rules

- Treat this directory as the reference home for telemetry transport and visibility topology.
- Keep collector, backend, scrape, datasource, and observability-stack compose configuration here.
- Do not move application-local runtime wiring or transport handler logic into this directory.
- Keep observability changes aligned with shared labels, trace correlation, and intent-event semantics defined by repository-level docs.

## Change Hints

- If a change alters repository-wide observability conventions or runtime semantics, update root `blueprint/` first.
- If a change stays within collector routing, backend config, or observability-stack composition, keep it local to this directory.
- If a change affects broader app-development startup wiring outside observability ownership, update `deployments/` instead.
