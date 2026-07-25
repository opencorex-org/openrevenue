# OpenRevenue

OpenRevenue is an open-source, configurable tax and revenue administration platform for governments, municipalities, and public-sector organizations. It is developed by the [OpenCorex open-source organization](https://github.com/opencorex-org) and is an early production-oriented foundation, not a finished revenue system.

> OpenRevenue is an independent [OpenCorex](https://opencorex.online) open-source project. It is not affiliated with Sri Lanka's Inland Revenue Department or any other government organization. All included tax data is fictional demonstration data and is not legal or tax advice.

## Vision and capabilities

The platform is designed for registration, filing, calculation, assessment, payments, append-only taxpayer ledgers, documents, notifications, audit, refunds, objections, compliance, reporting, and integrations. Country packs supply versioned forms, rules, languages, currencies, financial years, workflows, and integration adapters without putting national policy in the core.

The first vertical slice supports creating a taxpayer, approving a sample tax registration, drafting and validating a return, calculating a fictional 10% liability, submitting and assessing it, posting an assessment debit, receiving and allocating a payment, posting a payment credit, reading the ledger, recording audit events, and invoking a notification port.

## Applications and stack

- Go API, worker, and scheduler; Chi; PostgreSQL/pgx/sqlc migration-ready design; OpenAPI 3.1; slog; Prometheus; OpenTelemetry configuration.
- React/TypeScript/Vite portal workspace with TanStack Query and shared packages; the taxpayer portal is the initial UI shell.
- PostgreSQL, Redis, MinIO, Mailpit, Prometheus, Grafana, and OpenTelemetry Collector through Docker Compose.
- pnpm/Turborepo monorepo tooling and GitHub Actions foundations.

## Architecture

This is a domain-driven modular monolith. Modules own their data and expose application interfaces; domain code has no framework or persistence dependency. Cross-module asynchronous work uses a transactional outbox. Money uses signed 64-bit minor units and ISO-style currency codes. Returns retain form/rule versions. Ledger and audit tables reject updates and deletes.

See [architecture overview](docs/architecture/overview.md), [domain boundaries](docs/architecture/domain-boundaries.md), [country packs](docs/architecture/country-pack-architecture.md), and [security architecture](docs/architecture/security-architecture.md).

## Quick start

Requirements: Go 1.24+, Docker Compose, Node 22+, and pnpm 10+.

```sh
make setup
make test
make docker-up
make api
```

The API listens on `http://localhost:8080`; Mailpit is at `http://localhost:8025`. API routes require an `Authorization` header, and create-taxpayer requires `Idempotency-Key`. Run `make help` for commands.

## Repository map

`apps/` contains deployable processes and portals; `internal/` contains bounded contexts; `pkg/` contains narrow cross-cutting libraries; `web/` contains shared frontend packages; `contracts/` contains OpenAPI and event schemas; `database/` contains forward-only migrations; `country-packs/` contains jurisdiction configuration; `infrastructure/`, `docs/`, `tests/`, and `tools/` contain operational assets.

## Contributing, security, and roadmap

Read [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md), [GOVERNANCE.md](GOVERNANCE.md), and [ROADMAP.md](ROADMAP.md). Security reports must not be filed publicly. Licensed under Apache-2.0.
