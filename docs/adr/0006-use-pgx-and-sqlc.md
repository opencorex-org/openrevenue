# 0006: Use pgx and sqlc
**Status:** Accepted

## Context
SQL must remain visible, optimized, and statically typed.
## Decision
Use pgx pools and sqlc-generated repository code.
## Consequences
Queries are explicit and type safe; schema/query generation is a required build step.
## Alternatives considered
Runtime ORMs and database/sql-only access.
