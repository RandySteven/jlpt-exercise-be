package repositories

import (
	"context"
	"database/sql"

	repository_interfaces "github.com/RandySteven/go-cook/db"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/utils"
)

type (
	transaction struct {
		db *sql.DB
	}

	Transaction interface {
		RunInTx(ctx context.Context, txFunc func(ctx context.Context) *apperror.CustomError) (customErr *apperror.CustomError)
		RunInTxV2(ctx context.Context, steps ...stepFunc) (customErr *apperror.CustomError)
	}

	txCtxKey struct{}

	stepFunc func(ctx context.Context, params ...interface{}) (customErr *apperror.CustomError)
)

func (t *transaction) RunInTx(ctx context.Context, txFunc func(ctx context.Context) *apperror.CustomError) (customErr *apperror.CustomError) {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return apperror.DatabaseErrorErr()
	}
	txCtx := txToContext(ctx, tx)
	if customErr = txFunc(txCtx); customErr != nil {
		_ = tx.Rollback()
		return customErr
	}

	if err = tx.Commit(); err != nil {
		return apperror.NewCustomError(apperror.ErrInternalServer, `failed to commit tx`, err)
	}

	return nil
}

func (t *transaction) RunInTxV2(ctx context.Context, steps ...stepFunc) (customErr *apperror.CustomError) {
	err := utils.MutexLock(ctx, func(ctx context.Context) error {
		tx, err := t.db.BeginTx(ctx, nil)
		if err != nil {
			return apperror.NewCustomError(apperror.ErrInternalServer, `failed to begin tx`, err)
		}

		_ = txToContext(ctx, tx)
		for _, step := range steps {
			if err := step(ctx, tx); err != nil {
				_ = tx.Rollback()
				return apperror.NewCustomError(apperror.ErrInternalServer, `failed to run step`, err)
			}
		}

		if err = tx.Commit(); err != nil {
			return apperror.NewCustomError(apperror.ErrInternalServer, `failed to commit tx`, err)
		}
		return nil
	})
	if err != nil {
		return apperror.NewCustomError(apperror.ErrInternalServer, `failed to run in tx`, err)
	}

	return customErr
}

func txToContext(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

func txFromContext(ctx context.Context) *sql.Tx {
	tx, ok := ctx.Value(txCtxKey{}).(*sql.Tx)
	if ok {
		return tx
	}

	return nil
}

var _ Transaction = &transaction{}

func newTransaction(db *sql.DB) (*transaction, repository_interfaces.DBX) {
	return &transaction{
			db: db,
		}, func(ctx context.Context) repository_interfaces.Trigger {
			if tx := txFromContext(ctx); tx != nil {
				return tx
			}
			return db
		}
}
