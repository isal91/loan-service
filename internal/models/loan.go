package models

import (
	"time"
)

type LoanStatus string

const (
	StatusProposed  LoanStatus = "proposed"
	StatusApproved  LoanStatus = "approved"
	StatusInvested  LoanStatus = "invested"
	StatusDisbursed LoanStatus = "disbursed"
)

type Loan struct {
	ID                    int64      `db:"id"`
	LoanNumber            string     `db:"loan_number"`
	BorrowerID            int64      `db:"borrower_id"`
	Description           string     `db:"description"`
	PrincipalAmount       float64    `db:"principal_amount"`
	Rate                  float64    `db:"rate"`
	ROI                   float64    `db:"roi"`
	Status                LoanStatus `db:"status"`
	TotalInvested         float64    `db:"total_invested"`
	ApprovedAt            *time.Time `db:"approved_at"`
	ApprovedByEmployeeID  *string    `db:"approved_by_employee_id"`
	VisitProofURL         *string    `db:"visit_proof_url"`
	DisbursedAt           *time.Time `db:"disbursed_at"`
	DisbursedByEmployeeID *string    `db:"disbursed_by_employee_id"`
	BorrowerAgreementURL  *string    `db:"borrower_agreement_url"`
	CreatedAt             time.Time  `db:"created_at"`
}

func (s LoanStatus) IsValidTransition(to LoanStatus) bool {
	transitions := map[LoanStatus][]LoanStatus{
		StatusProposed: {StatusApproved},
		StatusApproved: {StatusInvested},
		StatusInvested: {StatusDisbursed},
	}

	allowed, ok := transitions[s]
	if !ok {
		return false
	}

	for _, v := range allowed {
		if v == to {
			return true
		}
	}
	return false
}
