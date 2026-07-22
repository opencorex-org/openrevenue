# Event guidelines

Events are immutable past-tense facts inside the standard envelope. Include identity, version, time, correlation, causation, actor, tenant, country, data, and metadata. Publish with the outbox, handle at least-once delivery idempotently, avoid sensitive payloads, and make same-version evolution additive.
