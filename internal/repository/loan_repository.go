package repository

import (
	"context"
	"database/sql"
	"loan-service/constant"
	"loan-service/internal/models"

	"github.com/jmoiron/sqlx"
)

type LoanRepository interface {
	Create(ctx context.Context, loan *models.Loan) error
	Update(ctx context.Context, loan *models.Loan) error
	GetByLoanNumber(ctx context.Context, loanNumber string) (*models.Loan, error)
	GetByLoanNumberWithLock(ctx context.Context, loanNumber string) (*models.Loan, error)
	GetByID(ctx context.Context, id int64) (*models.Loan, error)
	GetAllLoans(ctx context.Context) ([]models.Loan, error)
}

type loanRepository struct {
	db *sqlx.DB
}

func NewLoanRepository(db *sqlx.DB) LoanRepository {
	return &loanRepository{
		db: db,
	}
}

func (r *loanRepository) Create(ctx context.Context, l *models.Loan) error {
	query := `
		INSERT INTO loans (
			loan_number, borrower_id, description, principal_amount, 
			rate, roi, status, total_invested, created_at
		) VALUES (
			:loan_number, :borrower_id, :description, :principal_amount, 
			:rate, :roi, :status, :total_invested, :created_at
		) RETURNING id`

	model := models.Loan{
		LoanNumber:      l.LoanNumber,
		BorrowerID:      l.BorrowerID,
		Description:     l.Description,
		PrincipalAmount: l.PrincipalAmount,
		Rate:            l.Rate,
		ROI:             l.ROI,
		Status:          l.Status,
		TotalInvested:   l.TotalInvested,
		CreatedAt:       l.CreatedAt,
	}

	exec := GetExecutor(ctx, r.db)

	namedQuery, args, err := sqlx.Named(query, model)
	if err != nil {
		return err
	}
	namedQuery = r.db.Rebind(namedQuery)

	err = exec.QueryRowxContext(ctx, namedQuery, args...).Scan(&l.ID)
	return err
}

func (r *loanRepository) GetByID(ctx context.Context, id int64) (*models.Loan, error) {
	query := `
		SELECT 
			id, loan_number, borrower_id, description, principal_amount, 
			rate, roi, status, total_invested, approved_at, 
			approved_by_employee_id, visit_proof_url, disbursed_at, 
			disbursed_by_employee_id, borrower_agreement_url, created_at
		FROM loans WHERE id = $1`
	var m models.Loan
	exec := GetExecutor(ctx, r.db)

	err := sqlx.GetContext(ctx, exec, &m, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, constant.ErrNotFound
		}
		return nil, err
	}

	return r.toDomain(&m), nil
}

func (r *loanRepository) GetByLoanNumber(ctx context.Context, loanNumber string) (*models.Loan, error) {
	query := `
		SELECT 
			id, loan_number, borrower_id, description, principal_amount, 
			rate, roi, status, total_invested, approved_at, 
			approved_by_employee_id, visit_proof_url, disbursed_at, 
			disbursed_by_employee_id, borrower_agreement_url, created_at
		FROM loans WHERE loan_number = $1`
	var m models.Loan
	exec := GetExecutor(ctx, r.db)

	err := sqlx.GetContext(ctx, exec, &m, query, loanNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, constant.ErrNotFound
		}
		return nil, err
	}

	return r.toDomain(&m), nil
}

func (r *loanRepository) GetByLoanNumberWithLock(ctx context.Context, loanNumber string) (*models.Loan, error) {
	query := `
		SELECT 
			id, loan_number, borrower_id, description, principal_amount, 
			rate, roi, status, total_invested, approved_at, 
			approved_by_employee_id, visit_proof_url, disbursed_at, 
			disbursed_by_employee_id, borrower_agreement_url, created_at
		FROM loans WHERE loan_number = $1 FOR UPDATE`
	var m models.Loan
	exec := GetExecutor(ctx, r.db)

	err := sqlx.GetContext(ctx, exec, &m, query, loanNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, constant.ErrNotFound
		}
		return nil, err
	}

	return r.toDomain(&m), nil
}

func (r *loanRepository) Update(ctx context.Context, l *models.Loan) error {
	query := `
		UPDATE loans SET
			status = :status,
			approved_at = :approved_at,
			approved_by_employee_id = :approved_by_employee_id,
			visit_proof_url = :visit_proof_url,
			disbursed_at = :disbursed_at,
			disbursed_by_employee_id = :disbursed_by_employee_id,
			borrower_agreement_url = :borrower_agreement_url,
			total_invested = :total_invested
		WHERE id = :id`

	model := models.Loan{
		ID:                    l.ID,
		Status:                l.Status,
		ApprovedAt:            l.ApprovedAt,
		ApprovedByEmployeeID:  l.ApprovedByEmployeeID,
		VisitProofURL:         l.VisitProofURL,
		DisbursedAt:           l.DisbursedAt,
		DisbursedByEmployeeID: l.DisbursedByEmployeeID,
		BorrowerAgreementURL:  l.BorrowerAgreementURL,
		TotalInvested:         l.TotalInvested,
	}

	exec := GetExecutor(ctx, r.db)

	namedQuery, args, err := sqlx.Named(query, model)
	if err != nil {
		return err
	}
	namedQuery = r.db.Rebind(namedQuery)

	_, err = exec.ExecContext(ctx, namedQuery, args...)
	return err
}

func (r *loanRepository) GetAllLoans(ctx context.Context) ([]models.Loan, error) {
	query := `
		SELECT 
			id, loan_number, borrower_id, description, principal_amount, 
			rate, roi, status, total_invested, approved_at, 
			approved_by_employee_id, visit_proof_url, disbursed_at, 
			disbursed_by_employee_id, borrower_agreement_url, created_at
		FROM loans ORDER BY created_at DESC`
	var modelsList []models.Loan
	if err := r.db.SelectContext(ctx, &modelsList, query); err != nil {
		return nil, err
	}

	loans := make([]models.Loan, len(modelsList))
	for i, m := range modelsList {
		loans[i] = *r.toDomain(&m)
	}
	return loans, nil
}

func (r *loanRepository) toDomain(m *models.Loan) *models.Loan {
	return &models.Loan{
		ID:                    m.ID,
		LoanNumber:            m.LoanNumber,
		BorrowerID:            m.BorrowerID,
		Description:           m.Description,
		PrincipalAmount:       m.PrincipalAmount,
		Rate:                  m.Rate,
		ROI:                   m.ROI,
		Status:                m.Status,
		TotalInvested:         m.TotalInvested,
		ApprovedAt:            m.ApprovedAt,
		ApprovedByEmployeeID:  m.ApprovedByEmployeeID,
		VisitProofURL:         m.VisitProofURL,
		DisbursedAt:           m.DisbursedAt,
		DisbursedByEmployeeID: m.DisbursedByEmployeeID,
		BorrowerAgreementURL:  m.BorrowerAgreementURL,
		CreatedAt:             m.CreatedAt,
	}
}
