package models

type Pocket struct {
	ID                int64   `db:"id"`
	UserID            int64   `db:"user_id"`
	BalanceInvestable float64 `db:"balance_investable"`
	BalanceDisbursed  float64 `db:"balance_disbursed"`
}
