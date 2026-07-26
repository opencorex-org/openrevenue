-- name: ClaimOutboxRecords :many
WITH candidates AS (
    SELECT id
    FROM integration.outbox
    WHERE published_at IS NULL
      AND available_at <= now()
      AND (locked_until IS NULL OR locked_until <= now())
    ORDER BY occurred_at, id
    FOR UPDATE SKIP LOCKED
    LIMIT $1
)
UPDATE integration.outbox AS record
SET locked_by = $2,
    locked_until = now() + $3::interval
FROM candidates
WHERE record.id = candidates.id
RETURNING record.*;

-- name: MarkOutboxPublished :execrows
UPDATE integration.outbox
SET published_at = now(),
    locked_by = NULL,
    locked_until = NULL,
    last_error = NULL
WHERE id = $1
  AND locked_by = $2
  AND locked_until > now()
  AND published_at IS NULL;

-- name: RescheduleOutboxRecord :execrows
UPDATE integration.outbox
SET attempts = attempts + 1,
    available_at = now() + $3::interval,
    locked_by = NULL,
    locked_until = NULL,
    last_error = 'publication failed'
WHERE id = $1
  AND locked_by = $2
  AND locked_until > now()
  AND published_at IS NULL;
