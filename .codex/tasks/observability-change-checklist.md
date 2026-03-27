# Observability Change Checklist

Use this checklist when editing `platform/observability/` or observability-related local stack wiring.

## Read First

1. Read root `AGENTS.md`.
2. Read `blueprint/architecture/system-boundaries.md`.
3. Read `blueprint/conventions/observability-conventions.md`.
4. Read `platform/observability/README.md`.
5. Read `deployments/README.md` if the task may involve local stack wiring as well.

## Decide Scope

1. Edit `platform/observability/` for collector routing, backend config, dashboards, and observability stack composition.
2. Edit `deployments/` for broader local app stack wiring or operational wrappers.
3. Edit `blueprint/` first if observability semantics or repository-wide conventions change.
4. Edit `implementations/` only when services need code-level telemetry changes in addition to platform changes.

## Before Finishing

1. Check whether logs, traces, and metrics paths still match the intended topology.
2. Check whether shared labels or correlation fields need doc updates.
3. Check whether scenario or README guidance should change for validation steps.
