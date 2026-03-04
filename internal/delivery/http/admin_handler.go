package http

import (
	"loan-service/internal/usecase"
	"net/http"
)

type AdminHandler struct {
	loanUsecase usecase.LoanUsecase
}

func NewAdminHandler(u usecase.LoanUsecase) *AdminHandler {
	return &AdminHandler{loanUsecase: u}
}

func (h *AdminHandler) DebugGetLoans(w http.ResponseWriter, r *http.Request) {
	loans, err := h.loanUsecase.DebugGetLoans(r.Context())
	if err != nil {
		WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	WriteResponse(w, http.StatusOK, "success fetch loans", loans)
}

func (h *AdminHandler) DebugGetPockets(w http.ResponseWriter, r *http.Request) {
	pockets, err := h.loanUsecase.DebugGetPockets(r.Context())
	if err != nil {
		WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	WriteResponse(w, http.StatusOK, "success fetch pockets", pockets)
}

func (h *AdminHandler) DebugGetInvestments(w http.ResponseWriter, r *http.Request) {
	investments, err := h.loanUsecase.DebugGetInvestments(r.Context())
	if err != nil {
		WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	WriteResponse(w, http.StatusOK, "success fetch investments", investments)
}

func (h *AdminHandler) DebugGetLedger(w http.ResponseWriter, r *http.Request) {
	ledgers, err := h.loanUsecase.DebugGetLedgers(r.Context())
	if err != nil {
		WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
	WriteResponse(w, http.StatusOK, "success fetch ledgers", ledgers)
}
