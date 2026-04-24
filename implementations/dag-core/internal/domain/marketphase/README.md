Market Phase Domain

Purpose
- Resolve active market sessions (single phases) at a timestamp.
- Expose resolved single-phase windows for downstream aggregation and analysis.

Core Types
- `PhaseSpec`: canonical phase definition (market, timezone, local start/end, tags, notes).
- `ResolvedPhase`: runtime-resolved active single phase with local/UTC/JST boundaries.
- `PhaseBar`: persisted OHLCV aggregate for one resolved single phase window.

Main Flow
1) `DefaultPhaseSpecs()` provides baseline single-session specs.
2) `ResolveSinglePhases(at, specs)` selects active `single` phases at `at`.
3) `AggregatePhaseBar(phase, symbolID, ticks)` aggregates ticks for one resolved single phase window.

Single Phase Rules
- `at` must be non-zero; resolution is performed from UTC input.
- Each `PhaseSpec` requires `ID` and `Timezone`.
- Local interval must satisfy `end > start` on the same local day.
- Phase is active when `localStart <= localNow < localEnd`.
- Timezone loading is cached (`loadLocation`) to avoid repeated lookups.

Out Of Scope
- Derived transition / overlap signals such as `tokyo_london_transition`.
- Strategy-facing market context assembly across multiple single phases.
- Persistence or semantics for composite phases.

Responsibility Boundaries
- Domain (`marketphase`): single-phase taxonomy, time-window resolution, single-phase bar aggregation.
- Adjacent domain: composite market context or transition signals consume resolved singles but stay outside this package.
- Infrastructure layer: may persist or publish single-phase outputs, but does not own phase semantics.
