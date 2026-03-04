package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type contextKey string

const TxKey contextKey = "tx"

type transactionManager struct {
	db *sqlx.DB
}

func NewTransactionManager(db *sqlx.DB) TransactionManager {
	return &transactionManager{db: db}
}

func (m *transactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			// err is updated here and returned by WithinTransaction
			if commitErr := tx.Commit(); commitErr != nil {
				err = commitErr
			}
		}
	}()

	ctxWithTx := context.WithValue(ctx, TxKey, tx)
	err = fn(ctxWithTx)
	return err
}

func GetExecutor(ctx context.Context, db *sqlx.DB) sqlx.ExtContext {
	if tx, ok := ctx.Value(TxKey).(*sqlx.Tx); ok {
		return tx
	}
	return db
}
