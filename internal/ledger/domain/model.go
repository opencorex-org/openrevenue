package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/opencorex-org/openrevenue/pkg/id"
)

type TaxpayerTag struct{}
type RegistrationTag struct{}
type PeriodTag struct{}
type EntryTag struct{}

type TaxpayerID = id.ID[TaxpayerTag]
type RegistrationID = id.ID[RegistrationTag]
type PeriodID = id.ID[PeriodTag]
type EntryID = id.ID[EntryTag]

// Money stores minor currency units. Floating point values are never used.
type Money struct {
	Minor    int64  `json:"minor"`
	Currency string `json:"currency"`
}

func (m Money) Validate() error {
	if m.Minor < 0 {
		return errors.New("amount cannot be negative")
	}
	if len(m.Currency) != 3 {
		return errors.New("currency must be ISO 4217")
	}
	return nil
}
func (m Money) String() string {
	return fmt.Sprintf("%s %d.%02d", m.Currency, m.Minor/100, m.Minor%100)
}

type EntryType string

const (
	AssessmentDebit  EntryType = "ASSESSMENT_DEBIT"
	PaymentCredit    EntryType = "PAYMENT_CREDIT"
	PenaltyDebit     EntryType = "PENALTY_DEBIT"
	InterestDebit    EntryType = "INTEREST_DEBIT"
	RefundDebit      EntryType = "REFUND_DEBIT"
	AdjustmentDebit  EntryType = "ADJUSTMENT_DEBIT"
	AdjustmentCredit EntryType = "ADJUSTMENT_CREDIT"
	Reversal         EntryType = "REVERSAL"
)

type Entry struct {
	ID                EntryID           `json:"id"`
	TaxpayerID        TaxpayerID        `json:"taxpayerId"`
	TaxRegistrationID RegistrationID    `json:"taxRegistrationId"`
	TaxPeriodID       PeriodID          `json:"taxPeriodId"`
	Type              EntryType         `json:"entryType"`
	Debit             Money             `json:"debitAmount"`
	Credit            Money             `json:"creditAmount"`
	ReferenceType     string            `json:"referenceType"`
	ReferenceID       string            `json:"referenceId"`
	EffectiveDate     time.Time         `json:"effectiveDate"`
	PostedAt          time.Time         `json:"postedAt"`
	ReversalOf        *EntryID          `json:"reversalOf,omitempty"`
	CreatedBy         string            `json:"createdBy"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

func NewEntry(t EntryType, taxpayer TaxpayerID, registration RegistrationID, period PeriodID, amount Money, refType, refID, actor string, now time.Time) (Entry, error) {
	if err := amount.Validate(); err != nil {
		return Entry{}, err
	}
	if amount.Minor == 0 {
		return Entry{}, errors.New("ledger amount must be positive")
	}
	e := Entry{ID: id.New[EntryTag](), TaxpayerID: taxpayer, TaxRegistrationID: registration, TaxPeriodID: period, Type: t, ReferenceType: refType, ReferenceID: refID, EffectiveDate: now, PostedAt: now, CreatedBy: actor}
	if t == PaymentCredit || t == AdjustmentCredit {
		e.Credit = amount
		e.Debit = Money{Currency: amount.Currency}
	} else {
		e.Debit = amount
		e.Credit = Money{Currency: amount.Currency}
	}
	return e, nil
}

func NewReversal(original Entry, actor string, now time.Time) Entry {
	r := Entry{ID: id.New[EntryTag](), TaxpayerID: original.TaxpayerID, TaxRegistrationID: original.TaxRegistrationID, TaxPeriodID: original.TaxPeriodID, Type: Reversal, Debit: original.Credit, Credit: original.Debit, ReferenceType: "LEDGER_ENTRY", ReferenceID: original.ID.String(), EffectiveDate: now, PostedAt: now, CreatedBy: actor, ReversalOf: &original.ID}
	return r
}
