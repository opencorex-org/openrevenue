# 0005: Use PostgreSQL
**Status:** Accepted

## Context
Revenue records need transactional integrity, constraints, and durable auditability.
## Decision
Use PostgreSQL as the operational source of truth with logical schemas per context.
## Consequences
Strong consistency and rich indexing are available; migrations and operations need expertise.
## Alternatives considered
MySQL and document databases.
