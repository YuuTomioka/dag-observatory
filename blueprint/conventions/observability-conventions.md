# Observability Conventions

## Purpose

This document defines observability conventions for this parent repository.

## Principles

- observability is a structural requirement
- intent logs, traces, and metrics must be designed together
- telemetry naming and labels should be consistent across implementations
- cross-language propagation should preserve structural context

## Telemetry Roles

The repository distinguishes two log categories:

- intent logs: domain-meaningful events collected through OTel Logs
- environment logs: infrastructure and runtime logs collected through environment log pipelines

These should be searchable through a compatible label model even when they come from different collection paths.

## Label Conventions

The standard Loki-oriented labels are:

- `service`
- `env`
- `app`
- `instance`
- `job`

Event-specific identifiers such as `dag.run_id` and `dag.node_id` should stay as fields, not high-cardinality labels.

## Correlation Rules

- intent logs should include `event`
- intent logs should include `trace_id` and `span_id`
- cross-language task execution should preserve trace context
- application logs and intent logs should remain correlatable through shared identifiers

## Severity Guidance

- `INFO`: normal starts, finishes, and expected state changes
- `WARN`: degraded but non-fatal behavior such as skips
- `ERROR`: failed runs, failed nodes, or timeouts

## Metrics Guidance

Representative metrics should cover:

- HTTP request timing and size
- DAG run count and latency
- DAG node count, latency, and queue wait
- invalid intent-log emission count

Logs should expose attributes that can later support metrics derivation, such as `duration_ms`, `retry_count`, and `queue_wait_ms`.

## Migration Notes

This document is the durable home for repository-level observability conventions.
