package models

import "time"

type ActivityType string

const (
	ActivityInvestment   ActivityType = "investment"
	ActivityDisbursement ActivityType = "disbursement"
	ActivityRepayment    ActivityType = "repayment"
	ActivityTopUp        ActivityType = "topup"
	ActivityWithdrawal   ActivityType = "withdrawal"
)

type LedgerDirection string

const (
	DirectionCredit LedgerDirection = "CREDIT"
	DirectionDebit  LedgerDirection = "DEBIT"
)

type PocketLedger struct {
	ID           int64           `db:"id"`
	UserID       int64           `db:"user_id"`
	Amount       float64         `db:"amount"`
	Direction    LedgerDirection `db:"direction"`
	ActivityType ActivityType    `db:"activity_type"`
	ReferenceID  int64           `db:"reference_id"`
	CreatedAt    time.Time       `db:"created_at"`
}
