# Database guidelines

Each context owns one schema. Use explicit columns and constraints, `bigint` minor units for money, UUIDs for identity, UTC timestamps, and JSONB only for variable envelopes. Migrations are forward-only and must be tested from empty and prior release databases. Never update/delete ledger or audit facts; append corrections.
