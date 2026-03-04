package http

import (
	"encoding/json"
	"errors"
	"loan-service/constant"
	"loan-service/internal/pkg/logger"
	"net/http"
)

type Response struct {
	Code         string      `json:"code"`           // "success" or "failed"
	ErrorMessage string      `json:"message"`        // User-friendly message
	Data         interface{} `json:"data,omitempty"` // Response data
}

func WriteResponseWithCustomCode(w http.ResponseWriter, statusCode int, code string, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	res := Response{
		Code:         code,
		ErrorMessage: message,
		Data:         data,
	}

	json.NewEncoder(w).Encode(res)
}

func WriteResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	code := "success"
	if statusCode >= 400 {
		code = "failed"
	}
	WriteResponseWithCustomCode(w, statusCode, code, message, data)
}

func WriteError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	msg := err.Error()

	isDomainError := false
	domainErrors := []error{
		constant.ErrInvalidStatusTransition,
		constant.ErrInsufficientBalance,
		constant.ErrLoanFullyInvested,
		constant.ErrInvestmentAmountExceed,
		constant.ErrNotFound,
		constant.ErrLoanNotApproved,
		constant.ErrLoanNotInvested,
		constant.ErrDuplicateIdempotentKey,
	}

	for _, de := range domainErrors {
		if errors.Is(err, de) {
			isDomainError = true
			break
		}
	}

	code := "failed"
	// Default domain error code
	if isDomainError {
		// Override with specific codes for frontend handling
		if errors.Is(err, constant.ErrInsufficientBalance) {
			code = "insufficient_balance"
		} else if errors.Is(err, constant.ErrLoanFullyInvested) {
			code = "fully_funded"
		} else if errors.Is(err, constant.ErrDuplicateIdempotentKey) {
			code = "duplicate_request"
		} else if errors.Is(err, constant.ErrInvestmentAmountExceed) {
			code = "amount_exceed"
		} else if errors.Is(err, constant.ErrLoanNotApproved) {
			code = "not_approved"
		}
	}

	if !isDomainError && statusCode >= http.StatusInternalServerError {
		// Log the actual error for debugging
		logger.Error(r.Context(), "Internal Server Error", err)
		msg = "internal server error"
	}

	WriteResponseWithCustomCode(w, statusCode, code, msg, nil)
}
