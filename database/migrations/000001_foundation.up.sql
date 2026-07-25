CREATE SCHEMA IF NOT EXISTS identity; CREATE SCHEMA IF NOT EXISTS taxpayer;
CREATE SCHEMA IF NOT EXISTS registration; CREATE SCHEMA IF NOT EXISTS filing;
CREATE SCHEMA IF NOT EXISTS assessment; CREATE SCHEMA IF NOT EXISTS payment;
CREATE SCHEMA IF NOT EXISTS ledger; CREATE SCHEMA IF NOT EXISTS document;
CREATE SCHEMA IF NOT EXISTS notification; CREATE SCHEMA IF NOT EXISTS administration;
CREATE SCHEMA IF NOT EXISTS audit; CREATE SCHEMA IF NOT EXISTS integration;

CREATE TABLE taxpayer.taxpayers (id uuid PRIMARY KEY, tenant_id uuid NOT NULL, legal_name text NOT NULL, identifier text NOT NULL, created_at timestamptz NOT NULL DEFAULT now(), UNIQUE (tenant_id, identifier));
CREATE TABLE registration.tax_registrations (id uuid PRIMARY KEY, tenant_id uuid NOT NULL, taxpayer_id uuid NOT NULL, tax_type_code text NOT NULL, status text NOT NULL, created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE filing.tax_returns (id uuid PRIMARY KEY, tenant_id uuid NOT NULL, taxpayer_id uuid NOT NULL, registration_id uuid NOT NULL, period_id uuid NOT NULL, form_version text NOT NULL, rule_version text NOT NULL, status text NOT NULL, payload jsonb NOT NULL, submitted_at timestamptz);
CREATE TABLE assessment.assessments (id uuid PRIMARY KEY, tenant_id uuid NOT NULL, return_id uuid NOT NULL, amount_minor bigint NOT NULL CHECK (amount_minor >= 0), currency char(3) NOT NULL, created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE payment.payments (id uuid PRIMARY KEY, tenant_id uuid NOT NULL, taxpayer_id uuid NOT NULL, amount_minor bigint NOT NULL CHECK (amount_minor > 0), currency char(3) NOT NULL, external_reference text, created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE ledger.entries (id uuid PRIMARY KEY, tenant_id uuid NOT NULL, taxpayer_id uuid NOT NULL, tax_registration_id uuid NOT NULL, tax_period_id uuid NOT NULL, entry_type text NOT NULL, debit_minor bigint NOT NULL DEFAULT 0 CHECK (debit_minor >= 0), credit_minor bigint NOT NULL DEFAULT 0 CHECK (credit_minor >= 0), currency char(3) NOT NULL, reference_type text NOT NULL, reference_id uuid NOT NULL, effective_date date NOT NULL, posted_at timestamptz NOT NULL DEFAULT now(), reversal_of uuid REFERENCES ledger.entries(id), created_by text NOT NULL, metadata jsonb NOT NULL DEFAULT '{}', CHECK ((debit_minor > 0) <> (credit_minor > 0)));
CREATE OR REPLACE FUNCTION ledger.reject_mutation() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'ledger entries are append-only'; END $$;
CREATE TRIGGER ledger_entries_immutable BEFORE UPDATE OR DELETE ON ledger.entries FOR EACH ROW EXECUTE FUNCTION ledger.reject_mutation();
CREATE TABLE audit.events (id uuid PRIMARY KEY, tenant_id uuid NOT NULL, action text NOT NULL, actor text NOT NULL, resource_type text NOT NULL, resource_id text NOT NULL, occurred_at timestamptz NOT NULL, correlation_id text NOT NULL, metadata jsonb NOT NULL DEFAULT '{}');
CREATE TRIGGER audit_events_immutable BEFORE UPDATE OR DELETE ON audit.events FOR EACH ROW EXECUTE FUNCTION ledger.reject_mutation();
CREATE TABLE integration.outbox (id uuid PRIMARY KEY, aggregate_type text NOT NULL, aggregate_id text NOT NULL, event_type text NOT NULL, event_version integer NOT NULL, payload jsonb NOT NULL, occurred_at timestamptz NOT NULL, published_at timestamptz, attempts integer NOT NULL DEFAULT 0);
CREATE TABLE integration.idempotency_keys (tenant_id uuid NOT NULL, key text NOT NULL, operation text NOT NULL, response jsonb NOT NULL, created_at timestamptz NOT NULL DEFAULT now(), PRIMARY KEY (tenant_id, key, operation));
