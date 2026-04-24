Market Context Domain

Purpose
- Derive composite transition or overlap signals from resolved single phases.
- Build a compact strategy-facing market context without moving composite semantics into `marketphase`.

Core Types
- `SignalID`: identifier for composite market signals such as `tokyo_london_transition`.
- `Context`: summarized output (`ActivePhases`, `ActiveSignals`, `DominantMarkets`, `Tags`).

Main Flow
1) `marketphase.ResolveSinglePhases(at, specs)` resolves active single phases.
2) `ResolveSignals(activeSingles)` infers transition / overlap signals from those singles.
3) `Build(at, singles, signals)` aggregates phase IDs, signal IDs, markets, and tags.
4) `Resolve(at, specs)` is the end-to-end entrypoint for strategy-facing context.

Signal Rules
- `tokyo_london_transition`: active when `late_tokyo` is active and London local time is in `[07:00, 13:00)`.
- `london_newyork_overlap`: active when London core/late overlaps NY pre/core.
- `newyork_close_transition`: active when late NY remains active near close and other sessions are inactive.

Responsibility Boundaries
- Domain (`marketphase`): only single-phase taxonomy, time-window resolution, and single-phase bar aggregation.
- Domain (`marketcontext`): composite market signals and context assembly built from resolved singles.
- Infrastructure / application: consume `Context` but do not own transition logic.
