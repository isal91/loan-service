package response

type CreateLoanResponse struct {
	LoanNumber string `json:"loan_number"`
	Status     string `json:"status"`
}
