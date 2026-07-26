# Event architecture

Aggregate changes and an event envelope are committed to
`integration.outbox` in one transaction. A worker claims unpublished rows,
publishes versioned contracts, and records delivery attempts. Consumers use
event IDs as idempotency keys. Correlation and causation IDs preserve traces.
Event evolution is additive within a version; breaking changes create a new
version. See [return submission](../diagrams/tax-return-submission.md) and
[payment allocation](../diagrams/payment-allocation.md).

Domain commands use the platform `UnitOfWork` boundary. Domain state, its audit
fact, and the versioned event envelope are inserted before the same database
transaction commits. Any validation or insert failure rolls the complete
transaction back.

Outbox publishers claim ordered batches with `FOR UPDATE SKIP LOCKED` and a
bounded lease using `database/queries/outbox.sql`. Delivery is at least once;
consumers use tenant, jurisdiction, event ID, and version for idempotency.
Publish success and retry updates verify lease ownership so concurrent workers
cannot complete one another's records.

Audit events are append-only and tenant/jurisdiction scoped. Outbox rows may
change only in publication-control fields: lease, attempts, safe failure
category, and publication time. Event payloads remain unchanged.
