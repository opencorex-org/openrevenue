# 0002: Use a modular monolith
**Status:** Accepted

## Context
The domain is broad, but premature distribution adds failure modes and operational cost.
## Decision
Deploy a modular monolith with enforced bounded contexts and extractable contracts.
## Consequences
Local transactions remain possible; architecture tests and schema ownership are mandatory.
## Alternatives considered
Microservices and an unstructured monolith.
