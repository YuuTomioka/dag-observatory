# Observability Platform

This directory defines the reference operating environment for observability.

## Role

- collector, trace, log, and metrics reference topology
- local visibility and smoke-test guidance
- environment-level conventions that are not application code

In short:

- `platform/observability/` owns the observability platform definition
- `platform/observability/` does not own app-local runtime wiring outside observability concerns

## Current Asset Mapping

The current platform assets are split between permanent platform-owned assets and app-development composition assets.

They map conceptually as follows:

- `platform/observability/compose/docker-compose.observability.yml`: observability stack composition
- `platform/observability/otel-collector/otel-collector.yaml`: collector routing and processing
- `platform/observability/loki/loki.yaml`: log backend reference config
- `platform/observability/tempo/tempo.yaml`: trace backend reference config
- `platform/observability/prometheus/prometheus.yaml`: metrics scrape reference config
- `platform/observability/promtail/promtail.yaml`: environment log collection config
- `platform/observability/grafana/datasources.yaml`: operator-facing data source setup

Related app-development composition currently remains in:

- `deployments/compose/docker-compose.app.dev.yml`
- `deployments/compose/docker-compose.migrator.yml`

See `deployments/README.md` for the execution-wrapper role of those files.

When deciding where to edit:

- change `platform/observability/` if the task is about collector routing, backend config, or observability-stack composition
- change `deployments/` if the task is about starting the broader local app stack or running cross-cutting helpers

## Boundary Note

`docker-compose.migrator.yml` is not treated as an observability asset.

It remains under `deployments/compose/` because it is an operational execution wrapper for data migration, not part of the telemetry platform itself.

In other words:

- migration SQL belongs to `data/tsdb/`
- observability stack composition belongs to `platform/observability/`
- migrator execution remains a cross-cutting operational helper for now

## Reference Topology

The current reference topology is:

- application telemetry is sent to `otel-collector` over OTLP/HTTP
- logs are forwarded to Loki
- traces are forwarded to Tempo
- metrics are exposed for Prometheus scraping

Environment logs are collected separately and aligned to the same search model.

## Runtime Components

The current local reference environment includes:

- Loki for logs
- Tempo for traces
- Prometheus for metrics
- Grafana for operator visibility
- OTel Collector as telemetry routing hub
- Promtail for environment log collection

The broader app development environment also includes:

- TimescaleDB for time-series storage
- MinIO for artifact storage
- Redpanda for event transport

These are not all observability tools, but they form the practical runtime context in which observability is validated.

## Current As-Is Connection Notes

In the current local stack:

- application services send OTLP telemetry to `otel-collector`
- `otel-collector` forwards logs to Loki and traces to Tempo
- Prometheus scrapes collector-exposed metrics
- Grafana reads from Loki, Tempo, and Prometheus
- Promtail sends Docker environment logs directly to Loki

This means the current log path is intentionally split between:

- intent/application telemetry through OTel
- environment/runtime logs through Promtail

## Startup Outline

1. start observability infrastructure
2. start application services
3. confirm logs, traces, and metrics are visible
4. confirm shared labels and correlation fields are present

## Verification Focus

- intent logs are searchable with the expected labels
- `trace_id` and `span_id` are available for correlation
- runtime metrics are visible through Prometheus
- telemetry pipelines are reaching the collector and downstream backends

## Key Environment Variables

For the current Go reference implementation, the main observability-related variables are:

- `SERVICE_NAME`
- `SERVICE_NAMESPACE`
- `SERVICE_VERSION`
- `ENV`
- `HTTP_PORT`
- `OTEL_EXPORTER_OTLP_ENDPOINT`
- `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT`
- `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`
- `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`
- `APP_LOG_LEVEL`
- `APP_LOG_OUTPUT`

When endpoint-specific exporter variables are set, they override the base OTLP endpoint.

## Migration Notes

This document is now the durable home for observability platform guidance that previously lived in migration-era notes.

Existing runtime assets currently live under `deployments/`.
Only app-development wrappers and migrator wrappers still remain there.
