# Repository Purpose

## Positioning

This repository is a design template for derivative projects, not a finished product.

Its purpose is to preserve reusable structure for DAG execution, time-series data handling, asynchronous processing, and observability.

## In Scope

- Structural principles for DAG + time-series systems
- Reference implementations in Go and Python
- Reference operating environment for observability and local execution
- Reusable data design assets for time-series storage
- Scenarios and governance rules for repeated derivation

## Out of Scope

- Product-specific business requirements
- Environment-specific deployment decisions
- Feature completeness for a single product
- Temporary working notes as permanent documentation

## Inheritance to Derivatives

Derivative projects should inherit:

- system decomposition principles
- implementation boundary rules
- observability-first assumptions
- data design conventions

They should replace:

- domain workflows
- product-specific APIs
- product-specific deployment settings
- product-specific operating thresholds

## Runtime Terms

The repository uses the following core runtime terms across blueprint and implementations:

- Artifact: temporary data within one cycle
- State: data preserved across cycles
- Partition: the scope boundary for state
- Event: the normalized input unit consumed by the runtime

These terms should stay stable across derivative projects even when implementation details change.
