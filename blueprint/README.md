# Blueprint

This directory is the durable design source of truth for the repository.

## Role

- define repository-level structure and boundaries
- define durable flow, time, and observability assumptions
- define conventions that derivative projects should inherit
- record durable architectural decisions in ADRs

## Read Order

1. `identity/repository-purpose.md`
2. `architecture/system-boundaries.md`
3. `architecture/flow-and-time-model.md`
4. relevant files under `conventions/`
5. relevant files under `adr/` when past decisions matter

## What Belongs Here

- structural principles
- subsystem boundaries
- shared runtime concepts
- repository-wide conventions that should be inherited
- durable architectural rationale

## What Does Not Belong Here

- implementation-local API details
- local operating runbooks
- temporary task context
- generated artifacts

## Relationship To Other Directories

- use `governance/` for repository-wide operating rules and human-facing workflow surfaces
- use `implementations/` for implementation-local behavior and concrete API assets
- use `data/` for durable schema, migrations, and query assets
- use `platform/` for observability and operating environment reference assets
- use `scenarios/` for Japanese-facing validation and usage flows

## Update Rule

Update `blueprint/` first when a change affects:

- repository structure
- subsystem responsibility boundaries
- flow or time semantics
- inherited conventions for derivative projects
