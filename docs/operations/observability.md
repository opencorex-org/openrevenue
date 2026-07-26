# Observability

Emit structured redacted logs, Prometheus metrics, and OpenTelemetry traces with tenant-safe correlation. Monitor availability, latency, errors, saturation, outbox lag, notification failures, reconciliation exceptions, database health, and backup freshness. Alerts link to owned runbooks and actionable service-level objectives.

Monitor `openrevenue_domain_context_rejections_total` for ingress requests
rejected before application dispatch. Logs and traces may contain a correlation
ID but must not contain taxpayer identifiers, authorization headers, raw tenant
claims, or request payloads. Money mismatch and overflow failures are security-
and integrity-relevant application errors; alert on recurrence without recording
the underlying sensitive values.

The taxpayer-registration vertical slice exports low-cardinality counters for
successful taxpayer creation, registration submission, registration approval,
and safe request failures. Tenant identifiers, jurisdiction identifiers,
taxpayer identifiers, names, bearer tokens, and idempotency keys must never be
metric labels or log fields.

Operators should alert on a sustained increase in
`openrevenue_registration_vertical_slice_failures_total` relative to successful
submissions and approvals. Audit and integration events carry correlation and
causation identifiers for trace linkage without exposing taxpayer data.

Return lifecycle metrics report successful draft, validation, calculation,
submission, and amendment operations using fixed low-cardinality operation
labels. Alert on a sustained increase in
`openrevenue_return_lifecycle_failures_total`. Calculation explanations and
payload hashes may be retained with the return, but raw return lines and
validation inputs must not be emitted to logs, traces, or metric labels.

Financial-slice metrics report assessment posting, payment receipt, allocation,
and ledger reversal outcomes with fixed operation labels. Alert on
`openrevenue_financial_slice_failures_total`, stale allocation conflicts, and
imbalanced-posting rejections. Never emit payment references, source document
contents, taxpayer identifiers, amounts, or account balances as metric labels.

## Audit and outbox signals

Domain state, its append-only audit record, and its outbox envelope must be
written through one transaction. The transaction is rolled back if audit
metadata or event data contains a sensitive field. Allowed records contain
stable identifiers and lifecycle states only; names, taxpayer identifiers,
authorization values, tokens, request payloads, documents, and secrets are
rejected.

Publishers claim due records with a bounded lease. A record cannot be claimed by
two publishers while its lease is active. Successful publication records
`published_at`; failure increments `attempts`, stores only the redacted category
`publication failed`, releases the lease, and schedules a bounded retry.

The platform exports `openrevenue_outbox_pending`,
`openrevenue_outbox_failed`, and `openrevenue_outbox_oldest_age_seconds`.
The deployable Grafana dashboard is
`deploy/observability/openrevenue-dashboard.json`; Prometheus alert rules are in
`deploy/observability/openrevenue-alerts.yml`. Alerts cover publication
failures, backlog size and age, and critical API failure rates.

## Health, readiness, and trace correlation

`/health` reports process liveness only. `/ready` separately reports database,
migration, and broker readiness and returns HTTP 503 when any dependency is not
ready. Deployments must not route application traffic until readiness succeeds.

Ingress propagates or generates a correlation identifier and propagates a trace
identifier. Structured request logs contain the same correlation and trace IDs,
HTTP method, route, status, and duration. They never include headers, bodies,
tenant IDs, actor IDs, or domain identifiers.
