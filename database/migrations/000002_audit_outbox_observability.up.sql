ALTER TABLE audit.events
    ADD CONSTRAINT audit_events_metadata_object
    CHECK (jsonb_typeof(metadata) = 'object');

CREATE INDEX audit_events_tenant_time_idx
    ON audit.events (tenant_id, jurisdiction_code, occurred_at DESC, id);

ALTER TABLE integration.outbox
    ADD COLUMN available_at timestamptz NOT NULL DEFAULT now(),
    ADD COLUMN locked_by text,
    ADD COLUMN locked_until timestamptz,
    ADD COLUMN last_error text,
    ADD CONSTRAINT outbox_event_version_positive CHECK (event_version > 0),
    ADD CONSTRAINT outbox_payload_object CHECK (jsonb_typeof(payload) = 'object'),
    ADD CONSTRAINT outbox_attempts_nonnegative CHECK (attempts >= 0),
    ADD CONSTRAINT outbox_last_error_bounded CHECK (length(last_error) <= 160),
    ADD CONSTRAINT outbox_lock_complete
        CHECK ((locked_by IS NULL) = (locked_until IS NULL));

CREATE INDEX outbox_pending_idx
    ON integration.outbox (available_at, occurred_at)
    WHERE published_at IS NULL;

CREATE INDEX outbox_tenant_trace_idx
    ON integration.outbox (tenant_id, jurisdiction_code, correlation_id);
