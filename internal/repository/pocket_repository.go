package repository

import (
	"context"
	"loan-service/internal/models"

	"github.com/jmoiron/sqlx"
)

type PocketRepository interface {
	GetByUserID(ctx context.Context, userID int64) (*models.Pocket, error)
	GetByUserIDWithLock(ctx context.Context, userID int64) (*models.Pocket, error)
	Update(ctx context.Context, pocket *models.Pocket) error
	GetAllPockets(ctx context.Context) ([]models.Pocket, error)
}

type pocketRepository struct {
	db *sqlx.DB
}

func NewPocketRepository(db *sqlx.DB) PocketRepository {
	return &pocketRepository{db: db}
}

func (r *pocketRepository) GetByUserID(ctx context.Context, userID int64) (*models.Pocket, error) {
	query := `SELECT id, user_id, balance_investable, balance_disbursed, created_at FROM pockets WHERE user_id = $1`
	var m models.Pocket

	exec := GetExecutor(ctx, r.db)
	err := sqlx.GetContext(ctx, exec, &m, query, userID)
	if err != nil {
		return nil, err
	}

	return &models.Pocket{
		ID:                m.ID,
		UserID:            m.UserID,
		BalanceInvestable: m.BalanceInvestable,
		BalanceDisbursed:  m.BalanceDisbursed,
	}, nil
}

func (r *pocketRepository) GetByUserIDWithLock(ctx context.Context, userID int64) (*models.Pocket, error) {
	query := `SELECT id, user_id, balance_investable, balance_disbursed FROM pockets WHERE user_id = $1 FOR UPDATE`
	var m models.Pocket

	exec := GetExecutor(ctx, r.db)
	err := sqlx.GetContext(ctx, exec, &m, query, userID)
	if err != nil {
		return nil, err
	}

	return &models.Pocket{
		ID:                m.ID,
		UserID:            m.UserID,
		BalanceInvestable: m.BalanceInvestable,
		BalanceDisbursed:  m.BalanceDisbursed,
	}, nil
}

func (r *pocketRepository) Update(ctx context.Context, p *models.Pocket) error {
	query := `
		UPDATE pockets SET 
			balance_investable = :balance_investable,
			balance_disbursed = :balance_disbursed
		WHERE user_id = :user_id`

	model := models.Pocket{
		UserID:            p.UserID,
		BalanceInvestable: p.BalanceInvestable,
		BalanceDisbursed:  p.BalanceDisbursed,
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

func (r *pocketRepository) GetAllPockets(ctx context.Context) ([]models.Pocket, error) {
	query := `SELECT id, user_id, balance_investable, balance_disbursed FROM pockets`
	var pockets []models.Pocket
	if err := r.db.SelectContext(ctx, &pockets, query); err != nil {
		return nil, err
	}
	return pockets, nil
}
