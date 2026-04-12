# Task Documents

Task documents capture temporary working context.

They may also hold migration leftovers or derivative-planning notes until those decisions are either promoted into durable documents or retired.

Use this directory for shared temporary context.

Do not treat `.wrk/` as a repository note surface, and do not place runtime output or caches here. Local disposable artifacts belong in ignored `var/`.

## Required Fields

- background
- in-scope
- out-of-scope
- acceptance criteria
- result

## Rule

Task documents are not permanent structural documentation.

Durable conclusions must be promoted into:

- `blueprint/`
- `scenarios/`
- `governance/policies/`
- `blueprint/adr/`
