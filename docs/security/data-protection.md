# Data protection and classification

Collect only data required by an authorized revenue purpose. Production,
non-production, analytics, and support environments are separate; automated
tests and examples use fictional records only.

| Class and examples | Encryption | Access | Logs and telemetry | Retention and disposal |
| --- | --- | --- | --- | --- |
| Restricted secrets: tokens, passwords, private keys, signing/encryption material | Managed secret service and encrypted transport; never persisted in domain tables | Workload identity and named break-glass operators only | Never allowed; redact the entire value and nearby error detail | Rotate on schedule/event; revoke immediately; secret-service version policy |
| Restricted taxpayer data: identifiers, names, contact details, returns, payment references, documents | TLS 1.2+ in transit; provider-managed encryption at rest; field/token protection where required | Tenant scope plus role, purpose, ownership, assignment, and jurisdiction | No raw values, payloads, balances, or document contents | Statutory schedule and legal holds; verified deletion or cryptographic erasure |
| High-integrity financial records: assessments, allocations, ledger, reconciliation | TLS and encrypted storage/backups | Narrow command/query permissions; append-only correction and dual control where applicable | Stable record IDs only when operationally necessary; never amounts as labels | Statutory financial/audit retention; corrections by reversal |
| Confidential security evidence: audit events, IP/security signals, trace links, incident records | TLS and encrypted restricted stores | Auditor, security, and specifically authorized operators | Correlation IDs and stable categories; no credentials or taxpayer payloads | Evidence/incident schedule with legal holds and access review |
| Internal configuration: rule versions, deployment metadata, non-secret topology | TLS and encrypted repositories/stores | Authenticated staff and workloads by function | Approved low-cardinality version/status fields | Active life plus rollback/audit window |
| Public material: published forms, help, releases | Integrity-protected distribution | Public read, controlled publish | Normal service telemetry | Product and records schedule |

Backups inherit the highest classification they contain. Exports require purpose,
authorization, audit, encryption, expiry, and recipient verification. Support
tools must not copy production data into tickets or chat. Retention jobs are
tenant-aware, preserve statutory holds, record evidence without payloads, and
verify deletion from replicas, caches, object versions, and derived stores.
