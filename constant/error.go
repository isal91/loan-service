package constant

import "errors"

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrInsufficientBalance     = errors.New("insufficient balance")
	ErrLoanFullyInvested       = errors.New("loan is already fully invested")
	ErrInvestmentAmountExceed  = errors.New("investment amount exceeds principal")
	ErrNotFound                = errors.New("resource not found")
	ErrLoanNotApproved         = errors.New("loan must be in approved status to invest")
	ErrLoanNotInvested         = errors.New("loan must be fully invested to be disbursed")
	ErrDuplicateIdempotentKey  = errors.New("investment request already processed")
)
