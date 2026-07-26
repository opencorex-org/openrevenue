package application

import (
	"context"
	"time"

	foundation "github.com/opencorex-org/openrevenue/pkg/domain"
)

// UnitOfWork is the persistence boundary used by commands that change domain
// state. PostgreSQL implementations must execute the callback, audit insert,
// and outbox insert on the same database transaction and roll everything back
// when the callback fails.
type UnitOfWork interface {
	WithinTransaction(
		context.Context,
		foundation.Context,
		time.Time,
		func(TransactionWriter) error,
	) error
}

type TransactionWriter interface {
	Put(key, value string) error
	Audit(action, resourceType, resourceID string, metadata map[string]string) error
	Enqueue(eventType, aggregateType, aggregateID string, data map[string]string) error
}
