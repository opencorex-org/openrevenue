# 0001: Use a monorepo
**Status:** Accepted

## Context
Contracts, Go services, portals, packs, and deployment assets evolve together.
## Decision
Store them in one pnpm/Turborepo and Go-module repository.
## Consequences
Atomic changes and shared CI are easy; ownership and build caching must prevent coupling.
## Alternatives considered
Polyrepos were rejected until independent release ownership exists.
