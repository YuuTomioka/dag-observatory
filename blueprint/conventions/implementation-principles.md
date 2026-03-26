# Implementation Principles

## Purpose

This document defines implementation rules that apply across reference implementations.

## Principles

- structure takes priority over local convenience
- responsibility boundaries are more important than feature count
- code should explain the structure through clear module boundaries
- implementation-specific explanations should stay close to the implementation

## Workflow Compilation Rules

Workflow declarations should be compiled into executable form before runtime execution.

Compilation is responsible for validating at least:

- collisions between declared inputs and provided artifacts
- duplicate artifact providers
- duplicate state writers
- unresolved requirements
- graph cycles

Compiled output should retain both execution order and metadata needed for observation or dependency analysis.

## Node Execution Rules

Nodes should declare their structural contract rather than hiding it in implementation details.

The declaration surface should cover:

- required artifacts
- provided artifacts
- read state
- write state
- execution spec

## Execution Spec Rules

Execution characteristics should be explicit for both pure computation and external I/O nodes.

The execution spec should cover at least:

- determinism
- idempotency
- side effects
- timeout
- retry policy

Retry is allowed only when side effects remain safe under the declared idempotency rules.

If a retry happens, transaction state must be rebuilt from a fresh transaction rather than inherited implicitly.

## Transaction Rules

- state writes are staged and committed only on success
- failure causes rollback for the current attempt
- transaction scope is part of the execution model, not only a storage concern

## State Hash Rules

State hash is a comparison and audit aid, not a proof of correctness.

- hash targets should prefer deterministic state only
- dynamic timestamps, unordered maps, random values, and unstable external results should be excluded or normalized
- hashing strategy should remain replaceable by interface, not hard-coded into one store implementation

## Contract And Schema Rules

Contracts should be generated from the declared source of truth.

- generated artifacts must not be edited by hand
- generated artifacts remain review targets
- generation and verification should be part of standard automation

Per interface type, the source of truth must be explicit:

- REST: implementation source
- gRPC: proto source
- GraphQL: schema source
- WebSocket: typed message definitions

When multiple transports expose the same capability:

- they should converge on the same use case rather than reimplement business rules independently
- transport-local DTOs should remain transport-local
- generated contract artifacts should stay reviewable in version control

## Test And Review Rules

- boundary changes require blueprint review, not only implementation review
- generated contract diffs are specification review inputs
- implementation tests should confirm structure-preserving behavior, not only happy-path functionality

## Worker Integration Rules

When asynchronous workers are introduced:

- the worker should consume explicit runtime events rather than inventing its own hidden control plane
- routing from event type to handler should remain explicit
- event contracts should stay compatible with the shared runtime event model
- task execution logic should remain separate from transport, config, and serialization code

Workers may evolve from a single shared topic model to a split command/event model, but the structural principle stays the same: command handling and result reporting must remain explicit and reviewable.

## Transport Expansion Rules

- interface layers should stay limited to protocol handling, mapping, and routing
- domain and application layers should remain transport-agnostic
- read-oriented transports may project from shared views, but should not create competing state truths
- streaming and notification transports should be recoverable through snapshot-capable paths

## Migration Notes

This document is the durable home for repository-level implementation rules across reference implementations.
