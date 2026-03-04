package repository

import (
	"context"
	"strings"

	"loan-service/constant"
	"loan-service/internal/models"

	"github.com/jmoiron/sqlx"
)

type LedgerRepository interface {
	Create(ctx context.Context, ledger *models.PocketLedger) error
	GetAll(ctx context.Context) ([]models.PocketLedger, error)
}

type ledgerRepository struct {
	db *sqlx.DB
}

func NewLedgerRepository(db *sqlx.DB) LedgerRepository {
	return &ledgerRepository{db: db}
}

func (r *ledgerRepository) Create(ctx context.Context, l *models.PocketLedger) error {
	query := `
		INSERT INTO pocket_ledger (
			user_id, amount, direction, activity_type, reference_id, created_at
		) VALUES (
			:user_id, :amount, :direction, :activity_type, :reference_id, :created_at
		) RETURNING id`

	exec := GetExecutor(ctx, r.db)
	namedQuery, args, err := sqlx.Named(query, l)
	if err != nil {
		return err
	}
	namedQuery = r.db.Rebind(namedQuery)

	err = exec.QueryRowxContext(ctx, namedQuery, args...).Scan(&l.ID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") || strings.Contains(err.Error(), "23505") {
			return constant.ErrDuplicateIdempotentKey
		}
		return err
	}
	return nil
}

func (r *ledgerRepository) GetAll(ctx context.Context) ([]models.PocketLedger, error) {
	query := `SELECT id, user_id, amount, direction, activity_type, reference_id, created_at FROM pocket_ledger ORDER BY created_at DESC`
	var ledgers []models.PocketLedger
	if err := r.db.SelectContext(ctx, &ledgers, query); err != nil {
		return nil, err
	}
	return ledgers, nil
}
