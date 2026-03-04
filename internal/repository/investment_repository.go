package repository

import (
	"context"
	"database/sql"
	"loan-service/constant"
	"loan-service/internal/models"

	"github.com/jmoiron/sqlx"
)

type InvestmentRepository interface {
	Create(ctx context.Context, inv *models.Investment) error
	GetAll(ctx context.Context) ([]models.Investment, error)
	GetByLoanID(ctx context.Context, loanID int64) ([]models.Investment, error)
}

type investmentRepository struct {
	db *sqlx.DB
}

func NewInvestmentRepository(db *sqlx.DB) InvestmentRepository {
	return &investmentRepository{db: db}
}

func (r *investmentRepository) Create(ctx context.Context, inv *models.Investment) error {
	query := `
		INSERT INTO investments (
			loan_id, investor_id, amount, status, idempotent_key, created_at
		) VALUES (
			:loan_id, :investor_id, :amount, :status, :idempotent_key, :created_at
		) ON CONFLICT (investor_id, idempotent_key) DO NOTHING
		RETURNING id`

	exec := GetExecutor(ctx, r.db)

	namedQuery, args, err := sqlx.Named(query, inv)
	if err != nil {
		return err
	}
	namedQuery = r.db.Rebind(namedQuery)

	err = exec.QueryRowxContext(ctx, namedQuery, args...).Scan(&inv.ID)
	if err == sql.ErrNoRows {
		return constant.ErrDuplicateIdempotentKey
	}
	return err
}

func (r *investmentRepository) GetAll(ctx context.Context) ([]models.Investment, error) {
	query := `SELECT id, loan_id, investor_id, amount, status, idempotent_key, agreement_letter_url, created_at FROM investments ORDER BY created_at DESC`
	var investments []models.Investment
	if err := r.db.SelectContext(ctx, &investments, query); err != nil {
		return nil, err
	}
	return investments, nil
}

func (r *investmentRepository) GetByLoanID(ctx context.Context, loanID int64) ([]models.Investment, error) {
	query := `SELECT id, loan_id, investor_id, amount, status, idempotent_key, agreement_letter_url, created_at FROM investments WHERE loan_id = $1`
	var investments []models.Investment

	exec := GetExecutor(ctx, r.db)
	err := sqlx.SelectContext(ctx, exec, &investments, query, loanID)
	if err != nil {
		return nil, err
	}

	return investments, nil
}
