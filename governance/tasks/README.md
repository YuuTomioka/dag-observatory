# Task Documents

Task documents capture temporary working context.

They may also hold migration leftovers or derivative-planning notes until those decisions are either promoted into durable documents or retired.

Use this directory for shared temporary context.

Do not treat `.wrk/` as a repository note surface, and do not place runtime output or caches here. Local disposable artifacts belong in ignored `var/`.

## Naming Rule

- use lowercase kebab or snake case and keep one topic per file
- prefer a descriptive topic plus suffix, such as `topic_followups.md`, `topic_plan.md`, `topic_migration_plan.md`, or `topic_preparation.md`
- use the suffix to show the document role at a glance
- avoid generic names such as `memo.md`, `notes.md`, or `tmp.md`

## Required Fields

- background
- in-scope
- out-of-scope
- acceptance criteria
- result

Recommended additions when useful:

- references
- next actions
- promotion target

## Format Rule

- write task documents primarily in Japanese as shared human-facing working context
- keep one file focused on one working topic or one bounded temporary effort
- point to English source-of-truth documents instead of restating structure or technical truth here
- keep the top section stable so the file can be scanned quickly

## Lifecycle Rule

- create a task document when work needs temporary shared context
- update the same file while that topic remains active instead of creating fragmented note files
- promote durable conclusions into the proper permanent location when they stabilize
- retire the task document when the temporary working context is no longer needed

## Rule

Task documents are not permanent structural documentation.

Durable conclusions must be promoted into:

- `blueprint/`
- `scenarios/`
- `governance/policies/`
- `blueprint/adr/`
