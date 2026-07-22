# 0008: Use a transactional outbox
**Status:** Accepted

## Context
Database changes and distributed event publication cannot share a transaction.
## Decision
Commit versioned event envelopes beside aggregate changes and publish asynchronously.
## Consequences
No lost committed events; delivery is at least once and consumers must be idempotent.
## Alternatives considered
Direct publish and distributed transactions.
