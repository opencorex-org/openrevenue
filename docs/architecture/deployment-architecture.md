# Deployment architecture

Frontends are immutable static assets behind a CDN/WAF. API and workers scale horizontally; exactly one logical scheduler job runs through leader election or platform scheduling. PostgreSQL is the transactional system of record, Redis is non-authoritative, and object storage keeps documents. Health, readiness, metrics, traces, logs, backups, and secret rotation are mandatory operational capabilities. See [deployment diagram](../diagrams/deployment.md).
