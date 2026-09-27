package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// txBeginner is satisfied by *pgxpool.Pool and by pgx.Tx (where Begin opens a savepoint).
type txBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// PostgresTransactionManager implements the application's ITransactionManager with a pgx transaction carried in ctx.
type PostgresTransactionManager struct {
	db txBeginner
}

func NewPostgresTransactionManager(db txBeginner) *PostgresTransactionManager {
	return &PostgresTransactionManager{db: db}
}

// RunInTx begins a transaction, runs fn with a ctx carrying it, and commits when fn returns nil or rolls back otherwise.
func (m *PostgresTransactionManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}

	if err := fn(withTx(ctx, tx)); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

type txKey struct{}

// withTx returns a ctx carrying tx; repositories called with it run their queries on tx.
func withTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// querierFrom returns the transaction carried in ctx, or fallback when there is none.
func querierFrom(ctx context.Context, fallback querier) querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return fallback
}
