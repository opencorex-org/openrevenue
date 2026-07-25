package application

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	audit "github.com/opencorex-org/openrevenue/internal/audit/domain"
	filing "github.com/opencorex-org/openrevenue/internal/filing/domain"
	ledger "github.com/opencorex-org/openrevenue/internal/ledger/domain"
	"github.com/opencorex-org/openrevenue/pkg/id"
)

type TaxpayerTag struct{}
type RegistrationTag struct{}
type AssessmentTag struct{}
type PaymentTag struct{}
type Taxpayer struct {
	ID         id.ID[TaxpayerTag] `json:"id"`
	Name       string             `json:"name"`
	Identifier string             `json:"identifier"`
}
type Registration struct {
	ID         id.ID[RegistrationTag] `json:"id"`
	TaxpayerID string                 `json:"taxpayerId"`
	TaxType    string                 `json:"taxType"`
	Status     string                 `json:"status"`
}
type Assessment struct {
	ID       id.ID[AssessmentTag] `json:"id"`
	ReturnID string               `json:"returnId"`
	Amount   ledger.Money         `json:"amount"`
}
type Payment struct {
	ID          id.ID[PaymentTag] `json:"id"`
	TaxpayerID  string            `json:"taxpayerId"`
	Amount      ledger.Money      `json:"amount"`
	AllocatedTo string            `json:"allocatedTo"`
}
type Notification struct{ To, Subject, Body string }
type Notifier interface {
	Send(context.Context, Notification) error
}
type Service struct {
	mu            sync.RWMutex
	now           func() time.Time
	notifier      Notifier
	taxpayers     map[string]Taxpayer
	registrations map[string]Registration
	returns       map[string]filing.TaxReturn
	assessments   map[string]Assessment
	payments      map[string]Payment
	entries       []ledger.Entry
	audits        []audit.Event
	idempotency   map[string]any
}

func New(notifier Notifier) *Service {
	return &Service{now: time.Now, notifier: notifier, taxpayers: map[string]Taxpayer{}, registrations: map[string]Registration{}, returns: map[string]filing.TaxReturn{}, assessments: map[string]Assessment{}, payments: map[string]Payment{}, idempotency: map[string]any{}}
}
func (s *Service) record(action, actor, kind, rid, correlation string) {
	s.audits = append(s.audits, audit.New(action, actor, kind, rid, correlation, s.now()))
}
func (s *Service) CreateTaxpayer(name, identifier, actor, correlation, key string) (Taxpayer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.idempotency[key]; ok {
		return v.(Taxpayer), nil
	}
	if name == "" || identifier == "" {
		return Taxpayer{}, errors.New("name and identifier are required")
	}
	t := Taxpayer{ID: id.New[TaxpayerTag](), Name: name, Identifier: identifier}
	s.taxpayers[t.ID.String()] = t
	s.idempotency[key] = t
	s.record("TaxpayerRegistered", actor, "taxpayer", t.ID.String(), correlation)
	return t, nil
}
func (s *Service) Register(taxpayerID, taxType, actor, correlation string) (Registration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.taxpayers[taxpayerID]; !ok {
		return Registration{}, errors.New("taxpayer not found")
	}
	r := Registration{ID: id.New[RegistrationTag](), TaxpayerID: taxpayerID, TaxType: taxType, Status: "APPROVED"}
	s.registrations[r.ID.String()] = r
	s.record("TaxRegistrationApproved", actor, "registration", r.ID.String(), correlation)
	return r, nil
}
func (s *Service) DraftReturn(taxpayerID, registrationID, periodID string, lines []filing.Line, actor, correlation string) (filing.TaxReturn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.registrations[registrationID]
	if !ok || r.TaxpayerID != taxpayerID {
		return filing.TaxReturn{}, errors.New("registration not found")
	}
	tr := filing.New(taxpayerID, registrationID, periodID, lines)
	s.returns[tr.ID.String()] = tr
	s.record("ReturnCreated", actor, "return", tr.ID.String(), correlation)
	return tr, nil
}
func (s *Service) ValidateReturn(returnID, actor, correlation string) (filing.TaxReturn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.returns[returnID]
	if !ok {
		return r, errors.New("return not found")
	}
	if err := r.Validate(); err != nil {
		return r, err
	}
	s.returns[returnID] = r
	s.record("ReturnValidated", actor, "return", returnID, correlation)
	return r, nil
}
func (s *Service) SubmitAndAssess(ctx context.Context, returnID, actor, correlation string) (Assessment, error) {
	s.mu.Lock()
	r, ok := s.returns[returnID]
	if !ok {
		s.mu.Unlock()
		return Assessment{}, errors.New("return not found")
	}
	if err := r.Submit(s.now()); err != nil {
		s.mu.Unlock()
		return Assessment{}, err
	}
	var taxable int64
	for _, l := range r.Lines {
		taxable += l.AmountMinor
	}
	amount := ledger.Money{Minor: taxable / 10, Currency: "XCR"}
	a := Assessment{ID: id.New[AssessmentTag](), ReturnID: returnID, Amount: amount}
	e, err := ledger.NewEntry(ledger.AssessmentDebit, ledger.TaxpayerID(r.TaxpayerID), ledger.RegistrationID(r.RegistrationID), ledger.PeriodID(r.PeriodID), amount, "ASSESSMENT", a.ID.String(), actor, s.now())
	if err != nil {
		s.mu.Unlock()
		return Assessment{}, err
	}
	s.returns[returnID] = r
	s.assessments[a.ID.String()] = a
	s.entries = append(s.entries, e)
	s.record("ReturnSubmitted", actor, "return", returnID, correlation)
	s.record("AssessmentCreated", actor, "assessment", a.ID.String(), correlation)
	s.record("LedgerEntryPosted", actor, "ledger_entry", e.ID.String(), correlation)
	s.mu.Unlock()
	if s.notifier != nil {
		_ = s.notifier.Send(ctx, Notification{To: "demo@example.invalid", Subject: "Return submitted", Body: "Your fictional sample return was assessed."})
	}
	return a, nil
}
func (s *Service) RecordPayment(taxpayerID, assessmentID string, amount ledger.Money, actor, correlation string) (Payment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.assessments[assessmentID]
	if !ok {
		return Payment{}, errors.New("assessment not found")
	}
	r := s.returns[a.ReturnID]
	if r.TaxpayerID != taxpayerID {
		return Payment{}, errors.New("assessment does not belong to taxpayer")
	}
	if err := amount.Validate(); err != nil {
		return Payment{}, err
	}
	p := Payment{ID: id.New[PaymentTag](), TaxpayerID: taxpayerID, Amount: amount, AllocatedTo: assessmentID}
	e, err := ledger.NewEntry(ledger.PaymentCredit, ledger.TaxpayerID(r.TaxpayerID), ledger.RegistrationID(r.RegistrationID), ledger.PeriodID(r.PeriodID), amount, "PAYMENT", p.ID.String(), actor, s.now())
	if err != nil {
		return Payment{}, err
	}
	s.payments[p.ID.String()] = p
	s.entries = append(s.entries, e)
	s.record("PaymentReceived", actor, "payment", p.ID.String(), correlation)
	s.record("PaymentAllocated", actor, "assessment", assessmentID, correlation)
	s.record("LedgerEntryPosted", actor, "ledger_entry", e.ID.String(), correlation)
	return p, nil
}
func (s *Service) Ledger(taxpayerID string) []ledger.Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ledger.Entry, 0)
	for _, e := range s.entries {
		if e.TaxpayerID.String() == taxpayerID {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PostedAt.Before(out[j].PostedAt) })
	return out
}
func (s *Service) Audits() []audit.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]audit.Event(nil), s.audits...)
}
