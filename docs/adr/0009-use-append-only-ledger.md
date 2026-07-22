# 0009: Use an append-only ledger
**Status:** Accepted

## Context
Financial history must remain explainable and auditable.
## Decision
Forbid update/delete; correct with linked reversal and replacement entries using minor units.
## Consequences
History is complete; projections and reconciliation are needed for balances.
## Alternatives considered
Mutable balances and editable transactions.
