# Codex Adoption Preparation

## background

This task document translates the current Codex adoption proposal into repository-aligned preparation work.

The proposal direction is broadly compatible with the current repository shape:

- `blueprint/` remains the durable design source of truth
- `governance/` remains the durable operating-rule source of truth
- `implementations/`, `data/`, `platform/`, and `scenarios/` remain bounded working areas

The repository already has root and implementation-level `AGENTS.md` files, plus AI-local helper notes under `.codex/`.

Two proposal adjustments are required before durable rollout work:

- skill assets should be planned against repository-supported Codex skill locations rather than assuming `.codex/skills/`
- proposal artifacts should start as temporary shared task context under `governance/tasks/` until promoted

## in-scope

- define a phased task list for Codex adoption in this repository
- identify immediate preparation work that can be added without changing repository boundaries
- prepare initial Codex-local configuration and missing local `AGENTS.md` files
- consider repository document templates needed for Codex-assisted operating flow
- consider repository language policy for durable specs versus human-facing interaction surfaces
- capture corrections needed before later durable rollout work

## out-of-scope

- changing repository-wide structure or inheritance rules
- introducing new durable policy without separate promotion into `governance/policies/`
- implementing repository skills or subagent workflows in this task
- rewriting existing `blueprint/` architecture documents

## acceptance criteria

- a shared task list exists in `governance/tasks/`
- immediate preparation tasks are separated from later durable rollout tasks
- proposal adjustments and open decisions are explicit
- the document can be used as the working entrypoint for the next Codex adoption steps

## task list

1. Confirm proposal-to-repository mapping and record any required adjustments.
2. Update `.codex/config.toml` as the repository-scoped Codex working profile surface.
3. Update `data/tsdb/AGENTS.md` as the local read order and data-change rule surface.
4. Update `platform/observability/AGENTS.md` as the local read order and observability-change rule surface.
5. Review whether root `AGENTS.md` needs promotion-level changes after the first preparation pass.
6. Draft document template requirements for `governance/contracts/`, `governance/reviews/`, and any human-facing operation notes.
7. Define a proposed language policy:
   - durable specifications, structural design, and repository source-of-truth documents in English
   - human-facing request intake, handoff, and operating procedure surfaces in Japanese when they directly interface with repository users or operators
8. Promote the language policy into `governance/policies/documentation-policy.md` only after the boundary and template implications are reviewed.
9. Draft `governance/contracts/` templates for `change-request`, `implementation-task`, and `validation-report`.
10. Draft `governance/reviews/` templates only if review surfaces differ materially from existing task documents and `AGENTS.md` guidance.
11. Decide the repository location for Codex skills and align with current Codex repository skill discovery before scaffolding them.
12. Add initial skill candidates for change classification, blueprint-to-implementation translation, and scenario validation.
13. Add `scenarios/acceptance/` only when acceptance flows are concrete enough to avoid placeholder-only documentation.
14. Revisit whether subagents are needed after the above workflow surfaces stabilize.

## phased execution

### phase 1: immediate preparation

- update repository-scoped `.codex/config.toml`
- update local `AGENTS.md` files for `data/tsdb/` and `platform/observability/`
- keep all rollout planning in this task document until durable promotion is justified

### phase 2: workflow handoff surfaces

- define document template expectations before creating placeholder files
- decide which documents are operator-facing and therefore should be Japanese
- decide which documents remain durable repository specification and therefore stay in English
- when a document's role is ambiguous between specification and human-interface guidance, prefer Japanese
- add `governance/contracts/`
- define minimum fields for change request, implementation task, and validation report
- decide whether review templates belong under `governance/reviews/` or should stay implicit in existing review flows

### phase 3: repeatable Codex workflow assets

- scaffold repository skills in the correct repository skill location
- connect skills back to durable docs instead of duplicating design truth
- add acceptance scenarios only when they represent real validation flow

### phase 4: promotion review

- promote any durable operating rules into `governance/policies/`
- promote any structural conclusions into `blueprint/` or `blueprint/adr/`
- retire this task document when the temporary rollout context is no longer needed

## open questions

- whether a separate root-level durable document is needed for Codex operating modes beyond `AGENTS.md` and `.codex/config.toml`
- whether acceptance scenarios should be organized by subsystem or end-to-end flow

## working language direction

Current proposal for later promotion:

- `blueprint/`, ADRs, repository structure rules, and implementation/data/platform source-of-truth documents stay in English
- request intake, handoff forms, task communication templates, and operator-facing step-by-step procedures use Japanese when they are direct human-interface surfaces
- `scenarios/` is treated as a Japanese-facing validation and operation entry surface
- `governance/contracts/` and any later `governance/reviews/` templates should default to Japanese as human-interface documents
- documents with ambiguous boundaries between specification and human-interface guidance should default to Japanese
- mixed-language duplication should be avoided; each document should have one primary language based on its role
- if a Japanese human-facing document depends on durable repository truth, it should point back to the English source-of-truth document instead of redefining it

## result

Preparation started.

Completed in the current preparation pass:

- this document is now the shared rollout backlog
- `.codex/config.toml` was added as the repository-scoped Codex profile surface
- `data/tsdb/AGENTS.md` and `platform/observability/AGENTS.md` were added
- the language policy was promoted into `governance/policies/documentation-policy.md`
- initial templates were added under `governance/contracts/`
- root `AGENTS.md` was updated to include language and placement guidance for `governance/contracts/`, `scenarios/`, and `.agents/skills/`
- repository-scoped review templates are deferred; existing `governance/tasks/` documents and `/review` flow remain the default review surface
- the repository skill location was fixed at `.agents/skills/`
- initial skills were added for change classification, blueprint-to-implementation work, and scenario-based validation
- human-facing document aggregation was centered on `governance/`, with `governance/README.md` and `governance/guides/README.md` added as navigation surfaces
- naming, format, and lifecycle guidance was added for `governance/tasks/`, `governance/contracts/`, `governance/guides/`, and `scenarios/`

Next repository-safe moves are:

- add acceptance scenarios only when the validation flow is concrete enough
- revisit whether a separate root-level durable Codex operating document is needed beyond `AGENTS.md` and `.codex/config.toml`

Durable promotion work is intentionally deferred until the initial preparation pass is reviewed.
