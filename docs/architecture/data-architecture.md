# Data architecture

PostgreSQL uses one logical schema per bounded context. Tenant and country scope are explicit columns, with row-level security planned as defense in depth. Migrations are forward-only. Money is stored as `bigint` minor units plus a three-character currency. Submitted returns persist payload, form version, and rule version. Ledger and audit rows are append-only; corrections add reversals and replacements. Backups are encrypted and routinely restored in an isolated environment.
