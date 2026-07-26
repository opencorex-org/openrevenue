# Runbooks

Maintain procedures for API unavailability, database saturation/failover,
outbox backlog, reconciliation mismatch, notification failure, object-storage
outage, compromised credentials, suspicious access, migration failure, backup
failure, and country-pack rollback. Every runbook names signals, owner,
containment, diagnosis, recovery, verification, and escalation.

## Outbox publication

1. Confirm `openrevenue_outbox_failed`, pending count, and oldest age on the
   platform dashboard.
2. Check broker and database readiness without logging payloads.
3. Group failures by the stable publication category and correlation ID.
4. Restore the dependency or publisher capacity; leases expire automatically.
5. Verify backlog age and count return to baseline. Do not update outbox
   payloads manually or mark records published without downstream confirmation.

Escalate immediately when oldest age exceeds 15 minutes, the backlog continues
growing after recovery, or a critical event cannot publish.

## Critical API failures

1. Identify the affected fixed operation label and deployment version.
2. Follow correlation and trace IDs across redacted logs and traces.
3. Check database, migration, and broker readiness.
4. Roll back the deployment when the failure began with a release.
5. Preserve audit and outbox evidence; never copy request bodies, identifiers,
   tokens, or taxpayer data into the incident channel.
