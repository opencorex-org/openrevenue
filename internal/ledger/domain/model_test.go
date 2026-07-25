package domain

import (
	"github.com/opencorex-org/openrevenue/pkg/id"
	"testing"
	"time"
)

func TestReversalSwapsDebitAndCredit(t *testing.T) {
	e, err := NewEntry(AssessmentDebit, id.New[TaxpayerTag](), id.New[RegistrationTag](), id.New[PeriodTag](), Money{Minor: 2500, Currency: "XCR"}, "ASSESSMENT", "a", "u", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	r := NewReversal(e, "u", time.Now())
	if r.Credit.Minor != 2500 {
		t.Fatalf("credit = %d", r.Credit.Minor)
	}
	if r.ReversalOf == nil || e.ID != *r.ReversalOf {
		t.Fatal("reversal link missing")
	}
}
