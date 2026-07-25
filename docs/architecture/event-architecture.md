# Event architecture

Aggregate changes and an event envelope are committed to `integration.outbox` in one transaction. A worker claims unpublished rows, publishes versioned contracts, and records delivery attempts. Consumers use event IDs as idempotency keys. Correlation and causation IDs preserve traces. Event evolution is additive within a version; breaking changes create a new version. See [return submission](../diagrams/tax-return-submission.md) and [payment allocation](../diagrams/payment-allocation.md).
