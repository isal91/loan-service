package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"loan-service/constant"
	"loan-service/internal/models"
	"loan-service/internal/pkg/logger"
	"loan-service/internal/repository"

	"github.com/google/uuid"
)

type LoanUsecase interface {
	Propose(ctx context.Context, loan *models.Loan) error
	Approve(ctx context.Context, loanNumber string, employeeID string, visitProofURL string) error
	Invest(ctx context.Context, investorID int64, loanNumber string, amount float64, idempotentKey string) error
	Disburse(ctx context.Context, loanNumber string, employeeID string, agreementURL string) error
	// Debug Methods
	DebugGetLoans(ctx context.Context) ([]models.Loan, error)
	DebugGetPockets(ctx context.Context) ([]models.Pocket, error)
	DebugGetInvestments(ctx context.Context) ([]models.Investment, error)
	DebugGetLedgers(ctx context.Context) ([]models.PocketLedger, error)
}

type loanUsecase struct {
	loanRepo       repository.LoanRepository
	investmentRepo repository.InvestmentRepository
	ledgerRepo     repository.LedgerRepository
	pocketRepo     repository.PocketRepository
	txManager      repository.TransactionManager
}

func NewLoanUsecase(
	loanRepo repository.LoanRepository,
	investmentRepo repository.InvestmentRepository,
	ledgerRepo repository.LedgerRepository,
	pocketRepo repository.PocketRepository,
	txManager repository.TransactionManager,
) LoanUsecase {
	return &loanUsecase{
		loanRepo:       loanRepo,
		investmentRepo: investmentRepo,
		ledgerRepo:     ledgerRepo,
		pocketRepo:     pocketRepo,
		txManager:      txManager,
	}
}

func (u *loanUsecase) Propose(ctx context.Context, l *models.Loan) error {
	ctx, _ = logger.StartSpan(ctx, "Usecase.Propose")
	l.Status = models.StatusProposed
	l.TotalInvested = 0
	l.CreatedAt = time.Now()
	l.LoanNumber = "LN-" + uuid.New().String()[:8]

	if err := u.loanRepo.Create(ctx, l); err != nil {
		logger.Error(ctx, "failed to create loan", "error", err)
		return err
	}

	logger.Info(ctx, "loan proposed successfully", "loan_number", l.LoanNumber)
	return nil
}

func (u *loanUsecase) Approve(ctx context.Context, loanNumber string, employeeID string, visitProofURL string) error {
	ctx, _ = logger.StartSpan(ctx, "Usecase.Approve")
	loan, err := u.loanRepo.GetByLoanNumber(ctx, loanNumber)
	if err != nil {
		if errors.Is(err, constant.ErrNotFound) {
			logger.Warn(ctx, "loan not found for approval", "loan_number", loanNumber)
			return err
		}
		logger.Error(ctx, "failed to get loan for approval", "loan_number", loanNumber, "error", err)
		return err
	}

	if !loan.Status.IsValidTransition(models.StatusApproved) {
		logger.Warn(ctx, "invalid status transition for approval", "loan_number", loanNumber, "status", loan.Status)
		return constant.ErrInvalidStatusTransition
	}

	now := time.Now()
	loan.Status = models.StatusApproved
	loan.ApprovedAt = &now
	loan.ApprovedByEmployeeID = &employeeID
	loan.VisitProofURL = &visitProofURL

	if err := u.loanRepo.Update(ctx, loan); err != nil {
		logger.Error(ctx, "failed to update loan status to approved", "loan_number", loanNumber, "error", err)
		return err
	}

	logger.Info(ctx, "loan approved successfully", "loan_number", loanNumber)
	return nil
}

func (u *loanUsecase) Invest(ctx context.Context, investorID int64, loanNumber string, amount float64, idempotentKey string) error {
	ctx, _ = logger.StartSpan(ctx, "Usecase.Invest")
	return u.txManager.WithinTransaction(ctx, func(ctx context.Context) (err error) {
		ctx, _ = logger.StartSpan(ctx, "Usecase.Invest.Transaction")

		loan, err := u.loanRepo.GetByLoanNumberWithLock(ctx, loanNumber)
		if err != nil {
			logger.Error(ctx, "failed to get loan with lock for investment", "loan_number", loanNumber, "error", err)
			return err
		}

		if loan.Status != models.StatusApproved {
			logger.Warn(ctx, "attempt to invest in non-approved loan", "loan_number", loanNumber, "status", loan.Status)
			return constant.ErrLoanNotApproved
		}

		if loan.TotalInvested >= loan.PrincipalAmount {
			logger.Warn(ctx, "investment attempted on fully invested loan", "loan_number", loanNumber)
			return constant.ErrLoanFullyInvested
		}

		if loan.TotalInvested+amount > loan.PrincipalAmount {
			logger.Warn(ctx, "investment amount exceeds limit", "loan_number", loanNumber, "amount", amount)
			return constant.ErrInvestmentAmountExceed
		}

		pocket, err := u.pocketRepo.GetByUserIDWithLock(ctx, investorID)
		if err != nil {
			logger.Error(ctx, "failed to get investor pocket", "investor_id", investorID, "error", err)
			return err
		}

		if pocket.BalanceInvestable < amount {
			logger.Warn(ctx, "insufficient balance for investment", "investor_id", investorID, "amount", amount)
			return constant.ErrInsufficientBalance
		}

		pocket.BalanceInvestable -= amount
		if err := u.pocketRepo.Update(ctx, pocket); err != nil {
			logger.Error(ctx, "failed to update pocket balance", "investor_id", investorID, "error", err)
			return err
		}

		investment := &models.Investment{
			LoanID:        loan.ID,
			InvestorID:    investorID,
			Amount:        amount,
			Status:        string(models.InvStatusSuccess),
			IdempotentKey: idempotentKey,
			CreatedAt:     time.Now(),
		}
		if err := u.investmentRepo.Create(ctx, investment); err != nil {
			if errors.Is(err, constant.ErrDuplicateIdempotentKey) {
				logger.Info(ctx, "investment already processed (idempotent)", "investor_id", investorID, "idempotent_key", idempotentKey)
				return constant.ErrDuplicateIdempotentKey
			}
			logger.Error(ctx, "failed to create investment record", "loan_number", loanNumber, "error", err)
			return err
		}

		// Create Ledger Entry for Investor (DEBIT)
		ledger := &models.PocketLedger{
			UserID:       investorID,
			Amount:       amount,
			Direction:    models.DirectionDebit,
			ActivityType: models.ActivityInvestment,
			ReferenceID:  investment.ID,
			CreatedAt:    time.Now(),
		}
		if err := u.ledgerRepo.Create(ctx, ledger); err != nil {
			if errors.Is(err, constant.ErrDuplicateIdempotentKey) {
				logger.Info(ctx, "ledger entry already exists for investment (idempotent)", "investor_id", investorID, "loan_number", loanNumber)
				return constant.ErrDuplicateIdempotentKey
			}
			logger.Error(ctx, "failed to create ledger entry for investment", "investor_id", investorID, "error", err)
			return err
		}

		loan.TotalInvested += amount
		if loan.TotalInvested == loan.PrincipalAmount {
			loan.Status = models.StatusInvested

			// Get all investments for this loan
			investments, err := u.investmentRepo.GetByLoanID(ctx, loan.ID)
			if err != nil {
				logger.Error(ctx, "failed to get investments for sending agreement letter", "loan_number", loanNumber, "error", err)
				return err
			}

			// Known limitation: Since this is a take-home test, a message broker (e.g., Kafka) is not set up.
			// Using a simple goroutine as an asynchronous alternative to send agreement letters.
			defer func() {
				if err == nil {
					go func(invs []models.Investment, ln string) {
						// Use context.Background() since the HTTP req context might get cancelled
						bgCtx := context.Background()
						for _, inv := range invs {
							// Simulate generating and sending the agreement letter
							letterURL := fmt.Sprintf("https://cdn.example.com/agreements/%s/INV-%d.pdf", ln, inv.InvestorID)
							logger.Info(bgCtx, "Agreement letter generated and sent to investor asynchronously", "investor_id", inv.InvestorID, "loan_number", ln, "agreement_url", letterURL)
						}
					}(investments, loanNumber)
				}
			}()
		}
		if err := u.loanRepo.Update(ctx, loan); err != nil {
			logger.Error(ctx, "failed to update loan total invested", "loan_number", loanNumber, "error", err)
			return err
		}

		logger.Info(ctx, "investment processed successfully", "loan_number", loanNumber, "investor_id", investorID, "amount", amount)
		return nil
	})
}

func (u *loanUsecase) Disburse(ctx context.Context, loanNumber string, employeeID string, agreementURL string) error {
	ctx, _ = logger.StartSpan(ctx, "Usecase.Disburse")
	return u.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		ctx, _ = logger.StartSpan(ctx, "Usecase.Disburse.Transaction")
		loan, err := u.loanRepo.GetByLoanNumberWithLock(ctx, loanNumber)
		if err != nil {
			logger.Error(ctx, "failed to get loan for disbursement", "loan_number", loanNumber, "error", err)
			return err
		}

		if loan.Status == models.StatusDisbursed {
			logger.Info(ctx, "loan already disbursed (idempotent)", "loan_number", loanNumber)
			return constant.ErrDuplicateIdempotentKey // Reusing this for HTTP 409 mapping
		}

		if loan.Status != models.StatusInvested {
			logger.Warn(ctx, "attempt to disburse non-invested loan", "loan_number", loanNumber, "status", loan.Status)
			return constant.ErrLoanNotInvested
		}

		now := time.Now()
		loan.Status = models.StatusDisbursed
		loan.DisbursedAt = &now
		loan.DisbursedByEmployeeID = &employeeID
		loan.BorrowerAgreementURL = &agreementURL

		if err := u.loanRepo.Update(ctx, loan); err != nil {
			logger.Error(ctx, "failed to update loan to disbursed", "loan_number", loanNumber, "error", err)
			return err
		}

		pocket, err := u.pocketRepo.GetByUserIDWithLock(ctx, loan.BorrowerID)
		if err != nil {
			logger.Error(ctx, "failed to get borrower pocket for disbursement", "borrower_id", loan.BorrowerID, "error", err)
			return err
		}

		pocket.BalanceDisbursed += loan.PrincipalAmount
		if err := u.pocketRepo.Update(ctx, pocket); err != nil {
			logger.Error(ctx, "failed to update borrower pocket balance", "borrower_id", loan.BorrowerID, "error", err)
			return err
		}

		ledger := &models.PocketLedger{
			UserID:       loan.BorrowerID,
			Amount:       loan.PrincipalAmount,
			Direction:    models.DirectionCredit,
			ActivityType: models.ActivityDisbursement,
			ReferenceID:  loan.ID,
			CreatedAt:    time.Now(),
		}
		if err := u.ledgerRepo.Create(ctx, ledger); err != nil {
			if errors.Is(err, constant.ErrDuplicateIdempotentKey) {
				logger.Info(ctx, "ledger entry already exists for disbursement (idempotent)", "borrower_id", loan.BorrowerID, "loan_number", loanNumber)
				return constant.ErrDuplicateIdempotentKey
			}
			logger.Error(ctx, "failed to create ledger entry for disbursement", "borrower_id", loan.BorrowerID, "error", err)
			return err
		}

		logger.Info(ctx, "loan disbursed successfully", "loan_number", loanNumber)
		return nil
	})
}
