// Package application provides atomic audit and outbox transaction primitives.
package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	audit "github.com/opencorex-org/openrevenue/internal/audit/domain"
	event "github.com/opencorex-org/openrevenue/internal/integration/domain"
	foundation "github.com/opencorex-org/openrevenue/pkg/domain"
)

var (
	ErrSensitiveData  = errors.New("sensitive data is forbidden in audit and outbox records")
	ErrOutboxConflict = errors.New("outbox record is not available to this publisher")
)

type OutboxRecord struct {
	Event       event.Event
	AvailableAt time.Time
	PublishedAt *time.Time
	LockedBy    string
	LockedUntil *time.Time
	Attempts    uint32
	LastError   string
}

type Store struct {
	mu     sync.Mutex
	state  map[string]string
	audits []audit.Event
	outbox map[string]OutboxRecord
}

type Transaction struct {
	scope  foundation.Context
	now    time.Time
	state  map[string]string
	audits []audit.Event
	outbox map[string]OutboxRecord
}

func NewStore() *Store {
	return &Store{state: map[string]string{}, outbox: map[string]OutboxRecord{}}
}

func (s *Store) WithinTransaction(
	_ context.Context,
	scope foundation.Context,
	now time.Time,
	work func(TransactionWriter) error,
) error {
	return s.Transact(scope, now, func(tx *Transaction) error { return work(tx) })
}

func (s *Store) Transact(
	scope foundation.Context,
	now time.Time,
	work func(*Transaction) error,
) error {
	if err := scope.Validate(); err != nil {
		return err
	}
	if now.IsZero() || work == nil {
		return errors.New("transaction time and work are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx := &Transaction{
		scope: scope, now: now.UTC(),
		state: cloneMap(s.state), audits: append([]audit.Event(nil), s.audits...),
		outbox: cloneOutbox(s.outbox),
	}
	if err := work(tx); err != nil {
		return err
	}
	s.state, s.audits, s.outbox = tx.state, tx.audits, tx.outbox
	return nil
}

func (tx *Transaction) Put(key, value string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("state key is required")
	}
	tx.state[tx.scope.IsolationKey(key)] = value
	return nil
}

func (tx *Transaction) Audit(
	action, resourceType, resourceID string,
	metadata map[string]string,
) error {
	if err := rejectSensitive(metadata); err != nil {
		return err
	}
	record, err := audit.New(tx.scope, action, resourceType, resourceID, tx.now)
	if err != nil {
		return err
	}
	record.Metadata = cloneMap(metadata)
	tx.audits = append(tx.audits, record)
	return nil
}

func (tx *Transaction) Enqueue(
	eventType, aggregateType, aggregateID string,
	data map[string]string,
) error {
	if err := rejectSensitive(data); err != nil {
		return err
	}
	envelope, err := event.New(
		tx.scope, eventType, aggregateType, aggregateID, tx.now, cloneMap(data),
	)
	if err != nil {
		return err
	}
	tx.outbox[envelope.EventID.String()] = OutboxRecord{
		Event: envelope, AvailableAt: tx.now,
	}
	return nil
}

func (s *Store) Get(scope foundation.Context, key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.state[scope.IsolationKey(key)]
	return value, ok
}

func (s *Store) Audits(scope foundation.Context) []audit.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]audit.Event, 0)
	for _, record := range s.audits {
		if record.TenantID == scope.Tenant().String() &&
			record.Jurisdiction == scope.Jurisdiction().String() {
			record.Metadata = cloneMap(record.Metadata)
			result = append(result, record)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].OccurredAt.Before(result[j].OccurredAt)
	})
	return result
}

func (s *Store) Claim(now time.Time, publisher string, limit int, lease time.Duration) []OutboxRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	if publisher == "" || limit <= 0 || lease <= 0 {
		return nil
	}
	ids := make([]string, 0, len(s.outbox))
	for id := range s.outbox {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	claimed := make([]OutboxRecord, 0, limit)
	for _, id := range ids {
		record := s.outbox[id]
		if record.PublishedAt != nil || record.AvailableAt.After(now) ||
			(record.LockedUntil != nil && record.LockedUntil.After(now)) {
			continue
		}
		until := now.Add(lease).UTC()
		record.LockedBy, record.LockedUntil = publisher, &until
		s.outbox[id] = record
		claimed = append(claimed, cloneRecord(record))
		if len(claimed) == limit {
			break
		}
	}
	return claimed
}

func (s *Store) MarkPublished(id, publisher string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.outbox[id]
	if !ok || record.PublishedAt != nil || record.LockedBy != publisher ||
		record.LockedUntil == nil || !record.LockedUntil.After(now) {
		return ErrOutboxConflict
	}
	published := now.UTC()
	record.PublishedAt, record.LockedUntil, record.LockedBy = &published, nil, ""
	record.LastError = ""
	s.outbox[id] = record
	return nil
}

func (s *Store) MarkFailed(
	id, publisher string,
	now time.Time,
	retryAfter time.Duration,
	publicationError error,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.outbox[id]
	if !ok || record.PublishedAt != nil || record.LockedBy != publisher ||
		record.LockedUntil == nil || !record.LockedUntil.After(now) {
		return ErrOutboxConflict
	}
	record.Attempts++
	record.AvailableAt = now.Add(retryAfter).UTC()
	record.LockedBy, record.LockedUntil = "", nil
	record.LastError = safeError(publicationError)
	s.outbox[id] = record
	return nil
}

func (s *Store) OutboxStats(now time.Time) (pending, failed uint64, oldestAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, record := range s.outbox {
		if record.PublishedAt != nil {
			continue
		}
		pending++
		if record.Attempts > 0 {
			failed++
		}
		age := now.Sub(record.Event.OccurredAt)
		if age > oldestAge {
			oldestAge = age
		}
	}
	return
}

func rejectSensitive(values map[string]string) error {
	for key := range values {
		normalized := strings.ToLower(strings.ReplaceAll(key, "_", ""))
		for _, forbidden := range []string{
			"authorization", "token", "password", "secret", "taxpayeridentifier",
			"legalname", "payload", "document",
		} {
			if strings.Contains(normalized, forbidden) {
				return fmt.Errorf("%w: field %q", ErrSensitiveData, key)
			}
		}
	}
	return nil
}

func safeError(err error) string {
	if err == nil {
		return ""
	}
	return "publication failed"
}

func cloneMap(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func cloneRecord(source OutboxRecord) OutboxRecord {
	source.Event.Data = cloneMap(source.Event.Data)
	source.Event.Metadata = cloneMap(source.Event.Metadata)
	return source
}

func cloneOutbox(source map[string]OutboxRecord) map[string]OutboxRecord {
	result := make(map[string]OutboxRecord, len(source))
	for key, value := range source {
		result[key] = cloneRecord(value)
	}
	return result
}
