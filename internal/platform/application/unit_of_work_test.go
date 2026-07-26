package application

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	foundation "github.com/opencorex-org/openrevenue/pkg/domain"
)

func testScope(t *testing.T, tenant string) foundation.Context {
	t.Helper()
	tenantID, _ := foundation.NewTenantID(tenant)
	jurisdiction, _ := foundation.NewJurisdiction("LK")
	actor, _ := foundation.NewActor(foundation.ActorUser, "officer-1")
	correlation, _ := foundation.NewCorrelationID("request-1")
	scope, err := foundation.NewContext(tenantID, jurisdiction, actor, correlation)
	if err != nil {
		t.Fatal(err)
	}
	return scope
}

func TestTransactionCommitsStateAuditAndOutboxAtomically(t *testing.T) {
	store, scope := NewStore(), testScope(t, "tenant-one")
	now := time.Date(2026, 7, 26, 8, 0, 0, 0, time.UTC)
	err := store.Transact(scope, now, func(tx *Transaction) error {
		if err := tx.Put("return-1", "SUBMITTED"); err != nil {
			return err
		}
		if err := tx.Audit("ReturnSubmitted", "return", "return-1", nil); err != nil {
			return err
		}
		return tx.Enqueue("ReturnSubmitted", "return", "return-1", map[string]string{"status": "SUBMITTED"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if state, ok := store.Get(scope, "return-1"); !ok || state != "SUBMITTED" {
		t.Fatalf("state was not committed: %q %v", state, ok)
	}
	if len(store.Audits(scope)) != 1 {
		t.Fatal("audit was not committed")
	}
	if pending, _, _ := store.OutboxStats(now); pending != 1 {
		t.Fatalf("outbox was not committed: %d", pending)
	}
}

func TestTransactionRollbackAndSensitiveDataRejection(t *testing.T) {
	store, scope := NewStore(), testScope(t, "tenant-one")
	now := time.Now()
	err := store.Transact(scope, now, func(tx *Transaction) error {
		_ = tx.Put("return-1", "SUBMITTED")
		return tx.Enqueue("ReturnSubmitted", "return", "return-1", map[string]string{
			"authorizationToken": "must-not-leak",
		})
	})
	if !errors.Is(err, ErrSensitiveData) {
		t.Fatalf("expected sensitive-data rejection, got %v", err)
	}
	if _, ok := store.Get(scope, "return-1"); ok {
		t.Fatal("domain state committed without its outbox record")
	}
}

func TestAuditQueriesAreTenantScoped(t *testing.T) {
	store := NewStore()
	for _, tenant := range []string{"tenant-one", "tenant-two"} {
		scope := testScope(t, tenant)
		if err := store.Transact(scope, time.Now(), func(tx *Transaction) error {
			return tx.Audit("Read", "return", "return-1", nil)
		}); err != nil {
			t.Fatal(err)
		}
	}
	if got := store.Audits(testScope(t, "tenant-one")); len(got) != 1 ||
		got[0].TenantID != "tenant-one" {
		t.Fatalf("tenant audit isolation failed: %#v", got)
	}
	first := store.Audits(testScope(t, "tenant-one"))
	first[0].Metadata["modified"] = "outside"
	if _, exists := store.Audits(testScope(t, "tenant-one"))[0].Metadata["modified"]; exists {
		t.Fatal("returned audit metadata mutated the append-only record")
	}
}

func TestConcurrentClaimCannotDoublePublish(t *testing.T) {
	store, scope := NewStore(), testScope(t, "tenant-one")
	now := time.Now().UTC()
	if err := store.Transact(scope, now, func(tx *Transaction) error {
		return tx.Enqueue("ReturnSubmitted", "return", "return-1", nil)
	}); err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	wait.Add(2)
	claims := make(chan []OutboxRecord, 2)
	for _, publisher := range []string{"worker-1", "worker-2"} {
		go func(name string) {
			defer wait.Done()
			claims <- store.Claim(now, name, 1, time.Minute)
		}(publisher)
	}
	wait.Wait()
	close(claims)
	total := 0
	for claimed := range claims {
		total += len(claimed)
	}
	if total != 1 {
		t.Fatalf("expected one claim, got %d", total)
	}
}

func TestFailedPublicationIsRetriedWithoutLeakingControlCharacters(t *testing.T) {
	store, scope := NewStore(), testScope(t, "tenant-one")
	now := time.Now().UTC()
	_ = store.Transact(scope, now, func(tx *Transaction) error {
		return tx.Enqueue("ReturnSubmitted", "return", "return-1", nil)
	})
	claimed := store.Claim(now, "worker-1", 1, time.Minute)
	if len(claimed) != 1 {
		t.Fatal("record was not claimed")
	}
	if err := store.MarkFailed(
		claimed[0].Event.EventID.String(), "worker-1", now, time.Minute,
		errors.New("broker unavailable\nAuthorization: secret"),
	); err != nil {
		t.Fatal(err)
	}
	pending, failed, _ := store.OutboxStats(now.Add(time.Second))
	if pending != 1 || failed != 1 {
		t.Fatalf("unexpected stats pending=%d failed=%d", pending, failed)
	}
	retried := store.Claim(now.Add(time.Minute), "worker-2", 1, time.Minute)
	if len(retried) != 1 || retried[0].LastError != "publication failed" {
		t.Fatalf("failure was not safely redacted: %#v", retried)
	}
	if strings.Contains(retried[0].LastError, "secret") {
		t.Fatal("publication error leaked sensitive detail")
	}
	metrics := store.PrometheusMetrics(now.Add(time.Minute))
	for _, name := range []string{
		"openrevenue_outbox_pending 1",
		"openrevenue_outbox_failed 1",
		"openrevenue_outbox_oldest_age_seconds",
	} {
		if !strings.Contains(metrics, name) {
			t.Fatalf("missing metric %q: %s", name, metrics)
		}
	}
}
