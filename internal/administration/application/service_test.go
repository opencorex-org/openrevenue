package application_test

import (
	"context"
	"testing"

	app "github.com/opencorex-org/openrevenue/internal/administration/application"
	filing "github.com/opencorex-org/openrevenue/internal/filing/domain"
	ledger "github.com/opencorex-org/openrevenue/internal/ledger/domain"
)

type notifier struct{ sent int }

func (n *notifier) Send(context.Context, app.Notification) error { n.sent++; return nil }
func TestVerticalSlice(t *testing.T) {
	n := &notifier{}
	s := app.New(n)
	tp, err := s.CreateTaxpayer("Demo Cooperative", "DEMO-001", "admin", "corr", "key")
	if err != nil {
		t.Fatal(err)
	}
	reg, err := s.Register(tp.ID.String(), "SAMPLE_INCOME", "officer", "corr")
	if err != nil {
		t.Fatal(err)
	}
	ret, err := s.DraftReturn(tp.ID.String(), reg.ID.String(), "2026", []filing.Line{{Code: "GROSS", AmountMinor: 100_00}}, "taxpayer", "corr")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.ValidateReturn(ret.ID.String(), "taxpayer", "corr")
	if err != nil {
		t.Fatal(err)
	}
	a, err := s.SubmitAndAssess(context.Background(), ret.ID.String(), "taxpayer", "corr")
	if err != nil {
		t.Fatal(err)
	}
	if a.Amount.Minor != 10_00 {
		t.Fatalf("assessment = %d", a.Amount.Minor)
	}
	_, err = s.RecordPayment(tp.ID.String(), a.ID.String(), ledger.Money{Minor: 10_00, Currency: "XCR"}, "cashier", "corr")
	if err != nil {
		t.Fatal(err)
	}
	entries := s.Ledger(tp.ID.String())
	if len(entries) != 2 {
		t.Fatalf("entries = %d", len(entries))
	}
	if entries[0].Type != ledger.AssessmentDebit || entries[1].Type != ledger.PaymentCredit {
		t.Fatal("unexpected ledger entry types")
	}
	if len(s.Audits()) < 10 {
		t.Fatalf("audits = %d", len(s.Audits()))
	}
	if n.sent != 1 {
		t.Fatalf("notifications = %d", n.sent)
	}
}
