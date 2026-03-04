package usecase

import (
	"context"
	"loan-service/internal/models"
)

func (u *loanUsecase) DebugGetLoans(ctx context.Context) ([]models.Loan, error) {
	return u.loanRepo.GetAllLoans(ctx)
}

func (u *loanUsecase) DebugGetPockets(ctx context.Context) ([]models.Pocket, error) {
	return u.pocketRepo.GetAllPockets(ctx)
}

func (u *loanUsecase) DebugGetInvestments(ctx context.Context) ([]models.Investment, error) {
	return u.investmentRepo.GetAll(ctx)
}

func (u *loanUsecase) DebugGetLedgers(ctx context.Context) ([]models.PocketLedger, error) {
	return u.ledgerRepo.GetAll(ctx)
}
