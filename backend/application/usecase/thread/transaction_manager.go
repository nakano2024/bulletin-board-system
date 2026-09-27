package thread

import "context"

// ITransactionManager runs fn inside a single transaction: it commits when fn returns nil and rolls back otherwise.
// The transaction travels in the ctx passed to fn, so repositories called with that ctx join it.
type ITransactionManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
