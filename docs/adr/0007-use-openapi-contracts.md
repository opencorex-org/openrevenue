# 0007: Use OpenAPI contracts
**Status:** Accepted

## Context
Portals and integrations need stable, reviewable, typed APIs.
## Decision
Design versioned OpenAPI 3.1 contracts and generate Go/TypeScript boundaries.
## Consequences
Contract drift is testable; generation tooling must be pinned.
## Alternatives considered
Code-first REST and GraphQL.
