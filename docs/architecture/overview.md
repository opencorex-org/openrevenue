# Architecture overview

OpenRevenue is a deployable modular monolith with separate API, background worker, scheduler, and browser applications. Business capabilities are bounded contexts under `internal/`. A module owns its models and PostgreSQL schema; another module may call its application port or react to a versioned event, but may not query its tables. Domain packages depend only on the standard library and narrow domain types.

The synchronous path is HTTP → transport validation/authentication → application command/query → domain behavior → repository transaction. A transaction writes state, immutable audit facts, and outbox messages together. Workers publish and consume outbox records with idempotent handlers. [System context](../diagrams/system-context.md), [containers](../diagrams/container.md), and [module view](../diagrams/modular-monolith.md) show the boundaries.
