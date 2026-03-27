# Change Policy

## Classification Order

Before editing, classify the change in this order:

1. does it change repository structure, boundaries, or inheritance rules
2. does it change repository-wide documentation or change-handling rules
3. does it change representative usage or validation flow
4. does it stay within one implementation, one platform asset set, one data asset set, or one deployment wrapper

If the answer is higher in this list, update that layer first before editing lower layers.

## Update `blueprint/` When

- subsystem boundaries change
- flow or time modeling changes
- repository-wide conventions change
- a derivative project should inherit a new rule

Typical examples:

- moving responsibility between `platform/`, `deployments/`, `implementations/`, or `data/`
- changing event, artifact, state, partition, or observability semantics
- introducing a new reusable rule for derivative projects

## Update `governance/policies/` When

- change classification rules change
- documentation placement rules change
- AI and human contributors should follow a new repository-wide operating rule

## Update Implementations Only When

- the change stays within an existing boundary
- no repository-wide rule changes
- no scenario or governance meaning changes

Typical examples:

- transport handler changes inside `implementations/dag-core/`
- worker adapter or task behavior changes inside `implementations/worker-py/`
- implementation-local tests, contracts, or generated artifacts that follow existing rules

## Update `platform/` When

- observability platform topology or backend configuration changes
- collector, Loki, Tempo, Prometheus, Promtail, or Grafana reference assets change
- the operating reference environment changes without changing repository-wide structural rules

## Update `deployments/` When

- local app-development compose wiring changes
- migrator execution wrappers change
- helper containers or execution entrypoints change without changing platform ownership or data design truth

## Update `data/` When

- durable schema, migrations, or analysis queries change
- time-series design assets change independently of one implementation
- data conventions remain the same but the concrete SQL assets change

## Update Scenarios When

- the recommended entry flow changes
- the repository's representative use case changes
- operational validation steps materially change
