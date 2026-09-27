package postgres

import (
	"context"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

// PostgresTransactionManager implements the application's ITransactionManager with go-transaction-manager,
// which carries the pgx transaction in ctx; repositories pick it up via trmpgx.DefaultCtxGetter.
type PostgresTransactionManager struct {
	manager *manager.Manager
}

// NewPostgresTransactionManager takes db (*pgxpool.Pool in production) that begins the transactions.
func NewPostgresTransactionManager(db trmpgx.Transactional) *PostgresTransactionManager {
	return &PostgresTransactionManager{manager: manager.Must(trmpgx.NewDefaultFactory(db))}
}

// RunInTx runs fn in a transaction, committing when fn returns nil and rolling back otherwise.
func (m *PostgresTransactionManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.manager.Do(ctx, fn)
}
