package request

type CreateLoanRequest struct {
	BorrowerID      int64   `json:"borrower_id" validate:"required"` // For take-home test, this is a param; ideally extracted from JWT Token
	PrincipalAmount float64 `json:"principal_amount" validate:"required,gt=0"`
	Rate            float64 `json:"rate" validate:"required,gte=0"`
	ROI             float64 `json:"roi" validate:"required,gte=0"`
	Description     string  `json:"description" validate:"required"`
}

type ApproveLoanRequest struct {
	ApprovedByEmployeeID string `json:"approved_by_employee_id" validate:"required"`
	VisitProofURL        string `json:"visit_proof_url" validate:"required,url"`
}

type InvestLoanRequest struct {
	InvestorID    int64   `json:"investor_id" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	IdempotentKey string  `json:"idempotent_key" validate:"required"`
}

type DisburseLoanRequest struct {
	DisbursedByEmployeeID string `json:"disbursed_by_employee_id" validate:"required"`
	BorrowerAgreementURL  string `json:"borrower_agreement_url" validate:"required,url"`
}
