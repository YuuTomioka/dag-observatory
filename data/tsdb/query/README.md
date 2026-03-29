# TSDB Query Assets

This directory owns shared query SQL used by implementations.

## Naming Rules

- format: `NNNNNN_description.sql`
- `NNNNNN` is a zero-padded 6-digit sequence
- use lowercase snake case in `description`
- keep one bounded query topic per file when possible

## Current Layout

```txt
data/tsdb/query/
├─ README.md
├─ 000010_task_attempts.sql
├─ 000020_processed_events.sql
├─ 000030_outbox_events.sql
├─ 000040_artifacts.sql
├─ 000050_artifact_upload_sessions.sql
├─ 000060_db_backups.sql
├─ 000070_marketdata_symbol.sql
└─ 000080_marketdata_tick.sql
```

## Validation

```sh
make tsdb-query-lint
```
