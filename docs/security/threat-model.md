# Threat model

## Scope and security objectives

This model covers taxpayer, revenue officer, administrator, integration,
database, queue, document, and deployment paths. The primary objectives are
taxpayer confidentiality, tenant isolation, financial and ledger integrity,
service availability, audit evidence, and rule/version provenance.

## Assets and actors

| Asset | Required protection |
| --- | --- |
| Identity, contact, registration, returns, payments, and documents | Confidentiality, purpose limitation, integrity, statutory retention |
| Assessments, allocations, ledger, and reconciliation evidence | Strong integrity, append-only correction, availability |
| Credentials, signing keys, tokens, and encryption keys | Non-disclosure, least privilege, rotation, revocation |
| Audit events, outbox events, traces, and backups | Integrity, tenant isolation, bounded disclosure, recoverability |
| Source, dependencies, images, IaC, country packs, and generated artifacts | Provenance, review, reproducibility, vulnerability control |

Actors include taxpayers and authorized representatives, revenue officers,
administrators, auditors, workload identities, integration partners, operators,
support personnel, malicious outsiders, compromised accounts, and malicious or
coerced insiders.

## Trust boundaries

1. Public browser or integration client to edge/TLS termination.
2. Edge and identity provider to the API authentication boundary.
3. API commands to authorization, tenant, validation, and transaction
   boundaries.
4. Application workloads to PostgreSQL, queue/outbox publisher, object storage,
   mail, and telemetry collectors.
5. Document ingress to quarantine, malware scanning, and approved storage.
6. CI source and dependency inputs to build runners, registries, SBOMs, and
   deployment identities.
7. Operator access to production control planes, secret services, databases,
   backups, and incident evidence.

No identity, tenant, actor, forwarding, or TLS claim from an untrusted client is
authoritative until verified by the responsible boundary.

## Abuse cases and mitigations

| Path | Abuse case | Primary mitigations | Detection/evidence |
| --- | --- | --- | --- |
| Taxpayer | Credential theft, CSRF, object-reference tampering, bulk enumeration | OIDC, MFA where available, secure cookies/CSRF tokens, deny-by-default ownership checks, opaque IDs, rate limits | Authentication and authorization audit events; fixed-label failure metrics |
| Officer | Excessive search, assignment bypass, unauthorized adjustment or export | MFA, RBAC plus jurisdiction/assignment ABAC, step-up approval, export controls, least privilege | Query/export audit, anomaly alerts, periodic access review |
| Administrator | Privilege escalation, policy weakening, audit suppression | Separate administrative permissions, dual control for critical changes, immutable audit, workload separation | Configuration-change audit and high-severity alert |
| Integration | Token replay, forged tenant, schema abuse, duplicate commands | Workload identity, audience/scope validation, canonical tenant context, strict schemas, idempotency, request limits | Correlation/causation tracing and rejection counters |
| Database | SQL injection, cross-tenant query, direct ledger mutation, backup theft | Parameterized SQL, repository scoping, least-privilege roles, append-only triggers, TLS, encrypted storage/backups | Database audit, integrity reconciliation, restore tests |
| Queue/outbox | Event loss, replay, double publication, payload leakage, backlog denial | Atomic outbox, versioned envelopes, leased `SKIP LOCKED` claims, consumer idempotency, sensitive-field rejection | Backlog/failure/age metrics, trace IDs, immutable source record |
| Documents | Malware, content-type confusion, path traversal, public-object exposure | Size/type limits, generated object keys, quarantine, malware scanning, private buckets, short-lived access | Upload/download audit and scanner alerts |
| Deployment | Malicious dependency/action/image, secret exfiltration, unsafe IaC, rollback denial | SHA-pinned actions, lockfiles, SAST/dependency/secret/IaC/container scans, source and image SBOMs, non-root minimal image, workload identity | CI evidence, artifact provenance, registry and control-plane audit |
| Observability | Log injection or sensitive-data leakage | Structured logs, safe correlation formats, route templates, redaction, low-cardinality labels, restricted retention | Redaction tests and access audit |
| Availability | Resource exhaustion or dependency outage | Body limits, timeouts/rate limits at edge, bounded worker batches, readiness gates, capacity/backlog alerts, tested recovery | Health/readiness, SLO and saturation dashboards |

## Residual risk and review

External deployment remains blocked until identity, edge rate limiting,
production database roles, managed secrets, document malware scanning, backup
recovery, and alert delivery are verified in the target environment. Reassess
this model for new data classes, integrations, privileged actions, country-pack
execution, deployment paths, or material architecture changes.
