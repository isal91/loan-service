package http

import (
	"encoding/json"
	"errors"
	"loan-service/constant"
	"net/http"

	"loan-service/internal/delivery/http/request"
	"loan-service/internal/delivery/http/response"
	"loan-service/internal/models"
	"loan-service/internal/usecase"

	"github.com/go-playground/validator/v10"
)

type LoanHandler struct {
	loanUsecase usecase.LoanUsecase
}

func NewLoanHandler(u usecase.LoanUsecase) *LoanHandler {
	return &LoanHandler{
		loanUsecase: u,
	}
}

func (h *LoanHandler) Propose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, r, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	var req request.CreateLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	v := validator.New()
	if err := v.Struct(&req); err != nil {
		WriteError(w, r, errors.New("invalid request: "+err.Error()), http.StatusBadRequest)
		return
	}

	// Mapping DTO to Domain Entity
	loan := &models.Loan{
		BorrowerID:      req.BorrowerID,
		PrincipalAmount: req.PrincipalAmount,
		Rate:            req.Rate,
		ROI:             req.ROI,
		Description:     req.Description,
	}

	if err := h.loanUsecase.Propose(r.Context(), loan); err != nil {
		WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	// Mapping Domain to Response DTO
	res := response.CreateLoanResponse{
		LoanNumber: loan.LoanNumber,
		Status:     string(loan.Status),
	}

	WriteResponse(w, http.StatusCreated, "loan proposed successfully", res)
}

func (h *LoanHandler) Approve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, r, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	loanNumber := r.PathValue("loan_number")
	if loanNumber == "" {
		WriteError(w, r, errors.New("loan number is required"), http.StatusNotFound)
		return
	}

	var req request.ApproveLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	v := validator.New()
	if err := v.Struct(&req); err != nil {
		WriteError(w, r, errors.New("invalid request: "+err.Error()), http.StatusBadRequest)
		return
	}

	err := h.loanUsecase.Approve(r.Context(), loanNumber, req.ApprovedByEmployeeID, req.VisitProofURL)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constant.ErrInvalidStatusTransition) {
			status = http.StatusBadRequest
		}
		WriteError(w, r, err, status)
		return
	}

	WriteResponse(w, http.StatusOK, "loan approved successfully", nil)
}

func (h *LoanHandler) Invest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, r, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	loanNumber := r.PathValue("loan_number")
	if loanNumber == "" {
		WriteError(w, r, errors.New("loan number is required"), http.StatusNotFound)
		return
	}

	var req request.InvestLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	v := validator.New()
	if err := v.Struct(&req); err != nil {
		WriteError(w, r, errors.New("invalid request: "+err.Error()), http.StatusBadRequest)
		return
	}

	err := h.loanUsecase.Invest(r.Context(), req.InvestorID, loanNumber, req.Amount, req.IdempotentKey)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constant.ErrDuplicateIdempotentKey) {
			status = http.StatusConflict
		} else if errors.Is(err, constant.ErrInsufficientBalance) ||
			errors.Is(err, constant.ErrLoanFullyInvested) ||
			errors.Is(err, constant.ErrInvestmentAmountExceed) ||
			errors.Is(err, constant.ErrLoanNotApproved) {
			status = http.StatusBadRequest
		}
		WriteError(w, r, err, status)
		return
	}

	WriteResponse(w, http.StatusOK, "investment successful", nil)
}

func (h *LoanHandler) Disburse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, r, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	loanNumber := r.PathValue("loan_number")
	if loanNumber == "" {
		WriteError(w, r, errors.New("loan number is required"), http.StatusNotFound)
		return
	}

	var req request.DisburseLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	v := validator.New()
	if err := v.Struct(&req); err != nil {
		WriteError(w, r, errors.New("invalid request: "+err.Error()), http.StatusBadRequest)
		return
	}

	err := h.loanUsecase.Disburse(r.Context(), loanNumber, req.DisbursedByEmployeeID, req.BorrowerAgreementURL)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constant.ErrDuplicateIdempotentKey) {
			status = http.StatusConflict
		} else if errors.Is(err, constant.ErrLoanNotInvested) {
			status = http.StatusBadRequest
		}
		WriteError(w, r, err, status)
		return
	}

	WriteResponse(w, http.StatusOK, "loan disbursed successfully", nil)
}
