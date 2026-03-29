# TSDB Schema Assets

This directory owns durable schema-related SQL assets for `data/tsdb`.

## Scope

- `migrate/`: forward-only schema migrations
- `seed/`: optional initialization SQL for non-production or bootstrap needs

## Target Structure

```txt
data/tsdb/schema/
├─ README.md
├─ migrate/
│  ├─ 000001_init_extensions.sql
│  ├─ 000010_init_core_tables.sql
│  ├─ 000020_create_db_backups.sql
│  ├─ 000030_create_symbol.sql
│  └─ 000040_create_tick.sql
└─ seed/
   ├─ 000010_reference_symbols.sql
   └─ 000020_sample_ticks.sql
```

The exact migration contents may evolve, but naming and ordering rules should remain stable.

## Current State

`migrate/` currently follows the target naming format.

## Naming Rules

### `migrate/`

- format: `NNNNNN_description.sql`
- `NNNNNN` is a zero-padded 6-digit sequence
- use lowercase snake case in `description`
- keep migrations forward-only
- do not use `.up.sql` / `.down.sql` in this repository while the current migrator is filename-based

### `seed/`

- format: `NNNNNN_description.sql`
- `NNNNNN` is a zero-padded 6-digit sequence
- seed assets must be optional and replay-safe for local setup

## Why This Rule Exists

Current migrator behavior applies `*.sql` in lexical filename order and records applied filenames in `schema_migrations`.

Because of that behavior:

- mixed-width sequences such as `0001_*.sql` and `000010_*.sql` are error-prone
- renaming already-applied files can cause re-application risks

## Migration Plan

### Safe Plan (For Legacy Branches)

Use this when existing environments may already have applied current filenames.

1. Keep existing migration filenames as-is.
2. Start using the naming rules above for all new migration files only.
3. Optionally add a one-time linter/check that rejects non-conforming new filenames.
4. Document the policy in PRs until old files are naturally superseded.

This avoids conflicts with existing `schema_migrations(filename)` records.

Validation command:

```sh
make tsdb-schema-lint
```

### Clean Plan (Breaking / Reset Required)

Use this only when DB reset is acceptable for all target environments.

1. Backup required data before any reset.
2. Stop writers and migration runners.
3. Rename legacy files into the target naming format.
4. Recreate database (or clear schema and `schema_migrations`).
5. Re-apply migrations from scratch in lexical order.
6. Run application tests and migration smoke tests.
7. Update runbooks and onboarding docs.

Do not execute this plan against long-lived environments without explicit approval.
