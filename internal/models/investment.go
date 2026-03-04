package models

import "time"

type InvestmentStatus string

const (
	InvStatusSuccess InvestmentStatus = "invested"
)

type Investment struct {
	ID                 int64     `db:"id"`
	LoanID             int64     `db:"loan_id"`
	InvestorID         int64     `db:"investor_id"`
	Amount             float64   `db:"amount"`
	Status             string    `db:"status"`
	IdempotentKey      string    `db:"idempotent_key"`
	AgreementLetterURL *string   `db:"agreement_letter_url"`
	CreatedAt          time.Time `db:"created_at"`
}
